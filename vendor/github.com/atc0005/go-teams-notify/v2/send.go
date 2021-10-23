// Copyright 2020 Enrico Hoffmann
// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/go-teams-notify
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package goteamsnotify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// logger is a package logger that can be enabled from client code to allow
// logging output from this package when desired/needed for troubleshooting
var logger *log.Logger

// Known webhook URL prefixes for submitting messages to Microsoft Teams
const (
	WebhookURLOfficecomPrefix  = "https://outlook.office.com"
	WebhookURLOffice365Prefix  = "https://outlook.office365.com"
	WebhookURLOrgWebhookPrefix = "https://example.webhook.office.com"
)

// DisableWebhookURLValidation is a special keyword used to indicate to
// validation function(s) that webhook URL validation should be disabled.
//
// Deprecated: prefer using API.SkipWebhookURLValidationOnSend(bool) method instead
const DisableWebhookURLValidation string = "DISABLE_WEBHOOK_URL_VALIDATION"

// Regular Expression related constants that we can use to validate incoming
// webhook URLs provided by the user.
const (

	// DefaultWebhookURLValidationPattern is a minimal regex for matching known valid
	// webhook URL prefix patterns.
	DefaultWebhookURLValidationPattern = `^https:\/\/(?:.*\.webhook|outlook)\.office(?:365)?\.com`

	// Note: The regex allows for capital letters in the GUID patterns. This is
	// allowed based on light testing which shows that mixed case works and the
	// assumption that since Teams and Office 365 are Microsoft products case
	// would be ignored (e.g., Windows, IIS do not consider 'A' and 'a' to be
	// different).
	// webhookURLRegex           = `^https:\/\/(?:.*\.webhook|outlook)\.office(?:365)?\.com\/webhook(?:b2)?\/[-a-zA-Z0-9]{36}@[-a-zA-Z0-9]{36}\/IncomingWebhook\/[-a-zA-Z0-9]{32}\/[-a-zA-Z0-9]{36}$`

	// webhookURLSubURIWebhookPrefix         = "webhook"
	// webhookURLSubURIWebhookb2Prefix       = "webhookb2"
	// webhookURLOfficialDocsSampleURI       = "a1269812-6d10-44b1-abc5-b84f93580ba0@9e7b80c7-d1eb-4b52-8582-76f921e416d9/IncomingWebhook/3fdd6767bae44ac58e5995547d66a4e4/f332c8d9-3397-4ac5-957b-b8e3fc465a8c"
)

// ExpectedWebhookURLResponseText represents the expected response text
// provided by the remote webhook endpoint when submitting messages.
const ExpectedWebhookURLResponseText string = "1"

// DefaultWebhookSendTimeout specifies how long the message operation may take
// before it times out and is cancelled.
const DefaultWebhookSendTimeout = 5 * time.Second

// ErrWebhookURLUnexpected is returned when a provided webhook URL does
// not match a set of confirmed webhook URL patterns.
var ErrWebhookURLUnexpected = errors.New("webhook URL does not match one of expected patterns")

// ErrWebhookURLUnexpectedPrefix is returned when a provided webhook URL does
// not match a set of confirmed webhook URL prefixes.
//
// Deprecated: Use ErrWebhookURLUnexpected instead.
var ErrWebhookURLUnexpectedPrefix = ErrWebhookURLUnexpected

// ErrInvalidWebhookURLResponseText is returned when the remote webhook
// endpoint indicates via response text that a message submission was
// unsuccessful.
var ErrInvalidWebhookURLResponseText = errors.New("invalid webhook URL response text")

// API - interface of MS Teams notify
type API interface {
	Send(webhookURL string, webhookMessage MessageCard) error
	SendWithContext(ctx context.Context, webhookURL string, webhookMessage MessageCard) error
	SendWithRetry(ctx context.Context, webhookURL string, webhookMessage MessageCard, retries int, retriesDelay int) error
	SkipWebhookURLValidationOnSend(skip bool) API
	AddWebhookURLValidationPatterns(patterns ...string) API
	ValidateWebhook(webhookURL string) error
}

type teamsClient struct {
	httpClient                   *http.Client
	webhookURLValidationPatterns []string
	skipWebhookURLValidation     bool
}

func init() {
	// Disable logging output by default unless client code explicitly
	// requests it
	logger = log.New(os.Stderr, "[goteamsnotify] ", 0)
	logger.SetOutput(ioutil.Discard)
}

// EnableLogging enables logging output from this package. Output is muted by
// default unless explicitly requested (by calling this function).
func EnableLogging() {
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	logger.SetOutput(os.Stderr)
}

// DisableLogging reapplies default package-level logging settings of muting
// all logging output.
func DisableLogging() {
	logger.SetFlags(0)
	logger.SetOutput(ioutil.Discard)
}

// NewClient - create a brand new client for MS Teams notify
func NewClient() API {
	client := teamsClient{
		httpClient: &http.Client{
			// We're using a context instead of setting this directly
			// Timeout: DefaultWebhookSendTimeout,
		},
		skipWebhookURLValidation: false,
	}
	return &client
}

func (c *teamsClient) AddWebhookURLValidationPatterns(patterns ...string) API {
	c.webhookURLValidationPatterns = append(c.webhookURLValidationPatterns, patterns...)
	return c
}

// Send is a wrapper function around the SendWithContext method in order to
// provide backwards compatibility.
func (c teamsClient) Send(webhookURL string, webhookMessage MessageCard) error {
	// Create context that can be used to emulate existing timeout behavior.
	ctx, cancel := context.WithTimeout(context.Background(), DefaultWebhookSendTimeout)
	defer cancel()

	return c.SendWithContext(ctx, webhookURL, webhookMessage)
}

// SendWithContext posts a notification to the provided MS Teams webhook URL.
// The http client request honors the cancellation or timeout of the provided
// context.
func (c teamsClient) SendWithContext(ctx context.Context, webhookURL string, webhookMessage MessageCard) error {
	logger.Printf("SendWithContext: Webhook message received: %#v\n", webhookMessage)

	// optionally skip webhook validation
	if c.skipWebhookURLValidation {
		logger.Printf("SendWithContext: Webhook URL will not be validated: %#v\n", webhookURL)
	}

	// Validate input data
	if err := c.validateInput(webhookMessage, webhookURL); err != nil {
		return err
	}

	// prepare message
	webhookMessageByte, _ := json.Marshal(webhookMessage)
	webhookMessageBuffer := bytes.NewBuffer(webhookMessageByte)

	// Basic, unformatted JSON
	// logger.Printf("SendWithContext: %+v\n", string(webhookMessageByte))

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, webhookMessageByte, "", "\t"); err != nil {
		return err
	}
	logger.Printf("SendWithContext: Payload for Microsoft Teams: \n\n%v\n\n", prettyJSON.String())

	// prepare request (error not possible)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, webhookMessageBuffer)
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

	// do the request
	res, err := c.httpClient.Do(req)
	if err != nil {
		logger.Println(err)
		return err
	}

	if ctx.Err() != nil {
		logger.Println("SendWithContext: Context has expired after Do(req):", time.Now().Format("15:04:05"))
	}

	// Make sure that we close the response body once we're done with it
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	// Get the response body, then convert to string for use with extended
	// error messages
	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Println(err)
		return err
	}
	responseString := string(responseData)

	switch {
	// 400 Bad Response is likely an indicator that we failed to provide a
	// required field in our JSON payload. For example, when leaving out the
	// top level MessageCard Summary or Text field, the remote API returns
	// "Summary or Text is required." as a text string. We include that
	// response text in the error message that we return to the caller.
	case res.StatusCode >= 299:
		err = fmt.Errorf("error on notification: %v, %q", res.Status, responseString)
		logger.Println(err)
		return err

	// Microsoft Teams developers have indicated that a 200 status code is
	// insufficient to confirm that a message was successfully submitted.
	// Instead, clients should ensure that a specific response string was also
	// returned along with a 200 status code to confirm that a message was
	// sent successfully. Because there is a chance that unintentional
	// whitespace could be included, we explicitly strip it out.
	//
	// See atc0005/go-teams-notify#59 for more information.
	case responseString != strings.TrimSpace(ExpectedWebhookURLResponseText):

		err = fmt.Errorf(
			"got %q, expected %q: %w",
			responseString,
			ExpectedWebhookURLResponseText,
			ErrInvalidWebhookURLResponseText,
		)

		logger.Println(err)
		return err

	default:

		// log the response string
		logger.Printf("SendWithContext: Response string from Microsoft Teams API: %v\n", responseString)

		return nil
	}
}

// SendWithRetry is a wrapper function around the SendWithContext method in
// order to provide message retry support. The caller is responsible for
// provided the desired context timeout, the number of retries and retries
// delay.
func (c teamsClient) SendWithRetry(ctx context.Context, webhookURL string, webhookMessage MessageCard, retries int, retriesDelay int) error {
	var result error

	// initial attempt + number of specified retries
	attemptsAllowed := 1 + retries

	// attempt to send message to Microsoft Teams, retry specified number of
	// times before giving up
	for attempt := 1; attempt <= attemptsAllowed; attempt++ {
		// the result from the last attempt is returned to the caller
		result = c.SendWithContext(ctx, webhookURL, webhookMessage)

		switch {
		case result == nil:

			logger.Printf(
				"SendWithRetry: successfully sent message after %d of %d attempts\n",
				attempt,
				attemptsAllowed,
			)

			// No further retries needed
			return nil

		// While the context is passed to mstClient.SendWithContext and it
		// should ensure that it is respected, we check here explicitly in
		// order to return early in an effort to prevent undesired message
		// attempts
		case ctx.Err() != nil && result != nil:

			errMsg := fmt.Errorf(
				"SendWithRetry: context cancelled or expired: %v; "+
					"aborting message submission after %d of %d attempts: %w",
				ctx.Err().Error(),
				attempt,
				attemptsAllowed,
				result,
			)

			logger.Println(errMsg)
			return errMsg

		case result != nil:

			ourRetryDelay := time.Duration(retriesDelay) * time.Second

			logger.Printf(
				"SendWithRetry: Attempt %d of %d to send message failed: %v",
				attempt,
				attemptsAllowed,
				result,
			)

			// apply retry delay since our context hasn't been cancelled yet,
			// otherwise continue with the loop to allow context cancellation
			// handling logic to be applied
			logger.Printf(
				"SendWithRetry: Context not cancelled yet, applying retry delay of %v",
				ourRetryDelay,
			)
			time.Sleep(ourRetryDelay)
		}
	}

	return result
}

// SkipWebhookURLValidationOnSend allows the caller to optionally disable
// webhook URL validation.
func (c *teamsClient) SkipWebhookURLValidationOnSend(skip bool) API {
	c.skipWebhookURLValidation = skip
	return c
}

// validateInput verifies if the input parameters are valid
func (c teamsClient) validateInput(webhookMessage MessageCard, webhookURL string) error {
	// validate url
	if err := c.ValidateWebhook(webhookURL); err != nil {
		return err
	}

	// validate message
	return webhookMessage.Validate()
}

func (c teamsClient) ValidateWebhook(webhookURL string) error {
	if c.skipWebhookURLValidation || webhookURL == DisableWebhookURLValidation {
		return nil
	}

	u, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("unable to parse webhook URL %q: %w", webhookURL, err)
	}

	patterns := c.webhookURLValidationPatterns
	if len(patterns) == 0 {
		patterns = []string{DefaultWebhookURLValidationPattern}
	}

	// Return true if at least one pattern matches
	for _, pat := range patterns {
		matched, err := regexp.MatchString(pat, webhookURL)
		if err != nil {
			return err
		}
		if matched {
			return nil
		}
	}

	return fmt.Errorf("%w; got: %q, patterns: %s", ErrWebhookURLUnexpected, u.String(), strings.Join(patterns, ","))
}

// old deprecated helper functions --------------------------------------------------------------------------------------------------------------

// IsValidInput is a validation "wrapper" function. This function is intended
// to run current validation checks and offer easy extensibility for future
// validation requirements.
//
// Deprecated: use API.ValidateWebhook() and MessageCard.Validate()
// methods instead.
func IsValidInput(webhookMessage MessageCard, webhookURL string) (bool, error) {
	// validate url
	if valid, err := IsValidWebhookURL(webhookURL); !valid {
		return false, err
	}

	// validate message
	if valid, err := IsValidMessageCard(webhookMessage); !valid {
		return false, err
	}

	return true, nil
}

// IsValidWebhookURL performs validation checks on the webhook URL used to
// submit messages to Microsoft Teams.
//
// Deprecated: use API.ValidateWebhook() method instead.
func IsValidWebhookURL(webhookURL string) (bool, error) {
	c := teamsClient{}
	err := c.ValidateWebhook(webhookURL)
	return err == nil, err
}

// IsValidMessageCard performs validation/checks for known issues with
// MessardCard values.
//
// Deprecated: use MessageCard.Validate() instead.
func IsValidMessageCard(webhookMessage MessageCard) (bool, error) {
	err := webhookMessage.Validate()
	return err == nil, err
}

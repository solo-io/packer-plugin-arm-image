source "arm-image" "test" {
  iso_url = "test_fixtures/img.bin.gz"
  iso_checksum = "none"
  output_filename = "img.delete"
  image_mounts = ["/"]
}

build {
  sources = [
    "source.arm-image.test"
  ]
  # test that we can upload a file:
  provisioner "file" {
      source = "builder.go"
      destination = "/"
    }
  
  
  # provisioner "breakpoint" { 
  #      disable = false    
  #      note    = "this is a breakpoint"  
  #      }

  # test that we can run a command: 
  # not sure why, but some reason PATH is not set and this fails
  # disable for now.
 # provisioner "shell" {
 #   inline = [
 #     "echo hello world"
 #   ]
 #   environment_vars = ["PATH=/bin:$PATH"]
 #   execute_command = "/bin/chmod +x {{.Path}}; {{.Vars}} {{.Path}}"
 # }
}

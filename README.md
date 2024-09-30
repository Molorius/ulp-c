# ulp-c

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

ulp-c is a C compiler for the ESP32 ULP coprocessor. It is not yet functional.

# Status

This repository contains multiple compilers. C is compiled to Hlp, Hlp is compiled to assembly, assembly is compiled to binary. Below is the completion status:
- [X] Assembler
  - [X] Compiles
  - [X] Common code reduction 
- [ ] Hlp
  - [ ] Compiles
  - [ ] Optimizations
    - [ ] Register allocation
- [ ] C
  - [ ] Compiles
  - [ ] Examples

# Testing

The project can be tested on emulator with:
```sh
go test ./...
```

The project can be tested on hardware as well. In the next commands, change `PORT` to the desired USB port. First upload the test app to the esp32:
```sh
esptool.py --chip esp32 --port PORT --baud 460800 write_flash -z 0x1000 pkg/usb/test_app/test_app.bin
```

Then run all tests on emulator and hardware with:
```sh
ESP_PORT=PORT go test ./... -p 1
```



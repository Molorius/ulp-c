#!/usr/bin/env bash

idf.py build
esptool.py --chip esp32 merge_bin \
    --output test_app.bin \
    --target-offset 0x1000 \
    0x1000 build/bootloader/bootloader.bin \
    0x8000 build/partition_table/partition-table.bin \
    0x10000 build/test_app.bin

/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

// remove IDE warnings if esp-idf is not loaded
#ifdef ESP_PLATFORM

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

#include "stdio.h"
#include "string.h"
#include "mbedtls/base64.h"

#include "esp_log.h"
#include "esp_err.h"

#include "esp32/ulp.h"

#define ULP_START 0
#define TIMEOUT_MS 1000

#define ESP_ACK        0x0000 // the opposite device acknowledges the change
#define ESP_DONE       0x0001 // the ulp is done executing
#define ESP_PRINT_U16  0x0002 // the esp32 should print a u16 and a space
#define ESP_PRINT_CHAR 0x0003 // the esp32 should print a character

#define COMMAND_SIZE 1024*100 // this should be _plenty_ (under normal conditions)
#define BINARY_SIZE 8176 // this is max size of binary

static char command[COMMAND_SIZE] = { 0 };
static uint8_t binary[BINARY_SIZE];
static size_t bin_size;
static size_t command_index = 0;

typedef struct {
    uint32_t magic;
    uint16_t text_offset;
    uint16_t text_size;
    uint16_t data_size;
    uint16_t bss_size;
} u_header_t;

typedef struct {
    uint32_t *text;
    uint32_t *data;
    uint32_t *bss;
} u_mem_t;

// read a value from ulp .data
uint16_t u_read_data(u_mem_t *mem, size_t offset)
{
    uint32_t *addr;
    uint32_t val;

    addr = mem->data + offset;
    val = addr[0];
    return (uint16_t) (val & 0xFFFF);
}

// set a value in ulp .data
void u_set_data(u_mem_t *mem, size_t offset, uint32_t value)
{
    uint32_t *addr;

    addr = mem->data + offset;
    addr[0] = value;
}

// set up the memory offsets for later
void u_mem_setup(u_header_t header, u_mem_t *mem)
{
    const size_t load_addr = 0;

    mem->text = RTC_SLOW_MEM + load_addr;
    mem->data = mem->text + (header.text_size/4);
    mem->bss = mem->data + (header.data_size/4);
}

esp_err_t u_load_bin(u_header_t *header, const uint8_t *binary, size_t size)
{
    esp_err_t err;
    const uint32_t load_addr = 0;

    memcpy(header, binary, sizeof(u_header_t));
    err = ulp_load_binary(load_addr, binary, size);
    return err;
}

// take the ulp mutex
static void ulp_mutex_take(u_mem_t *mem)
{
    u_set_data(mem, 1, true); // flag[1] = true
    u_set_data(mem, 2, 0); // turn = 0
    while (u_read_data(mem, 0) && (u_read_data(mem, 2) == 0)) { // while flag[0] && turn == 0
        // busy wait
    }
}

// give the ulp mutex
static void ulp_mutex_give(u_mem_t *mem)
{
    u_set_data(mem, 1, false); // flag[1] = false
}


// read a character from input
static char character(void)
{
    char c;
    uint32_t start;

    for (;;) {
        start = esp_log_timestamp();
        while ((esp_log_timestamp()-start) < 10) {
            c = getchar();
            if (c != 0xFF)  {
                return c;
            }
        }
        vTaskDelay(1);
    }
}

// read a line
static esp_err_t readline(void)
{
    char c;

    command_index = 0;
    for(;;) {
        c = character();
        if (c == '\n') {
            putchar(' ');
            command[command_index] = 0;
            return 0;
        }
        if (c == '^') {
            return 1;
        }
        putchar(c);
        command[command_index] = c;
        command_index++;
        if (command_index > COMMAND_SIZE) { // too big, exit before memory gets too big
            return -1;
        }
    }
}

// erase the current input
static void erase_input(void)
{
    for (int i=0; i<COMMAND_SIZE; i++) {
        command[i] = 0;
    }
    command_index = 0;
}

// parse and decode the input
static esp_err_t parse(void)
{
    esp_err_t err;

    err = mbedtls_base64_decode(binary, BINARY_SIZE, &bin_size, (unsigned char*) command, strlen(command));
    return err;
}


// print error message
void printerr(void)
{
    printf(" ERR\n");
}

// print okay message
void printok(void)
{
    printf(" OK\n");
}

// erase the ulp memory
static void erase_ulp(void)
{
    int i;
    const int high = 8176 / sizeof(uint32_t);

    for (i=0; i<high; i++) {
        RTC_SLOW_MEM[i] = 0;
    }
}

// start the ulp
static esp_err_t ulp_start(void)
{
    esp_err_t err;
    uint32_t start;
    uint32_t fn;
    uint32_t param;
    u_header_t header;
    u_mem_t mem;

    err = u_load_bin(&header, binary, bin_size/sizeof(uint32_t));
    if (err) {
        return -10;
    }
    u_mem_setup(header, &mem);

    err = ulp_run(ULP_START);
    if (err) {
        return -12;
    }

    start = esp_log_timestamp();
    while ((esp_log_timestamp()-start) < TIMEOUT_MS) {
        ulp_mutex_take(&mem); // take the lock
        fn = u_read_data(&mem, 3); // read the function name
        param = u_read_data(&mem, 4); // read parameter
        u_set_data(&mem, 3, ESP_ACK); // write the acknowledgement
        ulp_mutex_give(&mem); // return lock
        switch (fn) {
            case ESP_ACK:
                vTaskDelay(1);
                break;
            case ESP_DONE:
                return 0;
            case ESP_PRINT_U16:
                printf("%lu ", param);
                break;
            case ESP_PRINT_CHAR:
                printf("%c", (char) param);
                break;
            default:
                return -1;
        }
    }

    return -2;
}

void app_main(void)
{
    esp_err_t err;

    for (;;) {
        erase_input();
        err = readline();
        if (err < 0) {
            printerr();
            continue;
        }
        else if (err == 1) {
            continue;
        }
        if (strlen(command) == 0) {
            printok();
            continue;
        }
        err = parse();
        if (err) {
            printf("decoding %i ERR\n", err);
            continue;
        }
        erase_ulp();
        err = ulp_start();
        if (err) {
            printf("ulp %i ERR\n", err);
            continue;
        }
        
        vTaskDelay(1);
        printok();
    }
}

#endif // #ifndef ESP_PLATFORM

//go:build darwin

#import <Foundation/Foundation.h>
#import <IOKit/IOKitLib.h>
#include <stdlib.h>

typedef struct {
    unsigned long long read_bytes;
    unsigned long long write_bytes;
} DiskStats;

DiskStats pollDiskStats() {
    DiskStats stats = {0};

    io_iterator_t iter;
    kern_return_t kr = IOServiceGetMatchingServices(
        kIOMainPortDefault,
        IOServiceMatching("IOBlockStorageDriver"),
        &iter
    );
    if (kr != KERN_SUCCESS) {
        return stats;
    }

    io_service_t entry;
    while ((entry = IOIteratorNext(iter)) != IO_OBJECT_NULL) {
        CFMutableDictionaryRef props = NULL;
        kr = IORegistryEntryCreateCFProperties(entry, &props, kCFAllocatorDefault, 0);
        IOObjectRelease(entry);

        if (kr != KERN_SUCCESS || props == NULL) continue;

        CFDictionaryRef statistics = CFDictionaryGetValue(props, CFSTR("Statistics"));
        if (statistics != NULL) {
            CFNumberRef readRef = CFDictionaryGetValue(statistics, CFSTR("Bytes (Read)"));
            CFNumberRef writeRef = CFDictionaryGetValue(statistics, CFSTR("Bytes (Write)"));

            long long val;
            if (readRef && CFNumberGetValue(readRef, kCFNumberLongLongType, &val)) {
                stats.read_bytes += (unsigned long long)val;
            }
            if (writeRef && CFNumberGetValue(writeRef, kCFNumberLongLongType, &val)) {
                stats.write_bytes += (unsigned long long)val;
            }
        }

        CFRelease(props);
    }

    IOObjectRelease(iter);
    return stats;
}

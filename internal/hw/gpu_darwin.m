//go:build darwin

#import <Foundation/Foundation.h>
#import <IOKit/IOKitLib.h>
#include <stdlib.h>
#include <string.h>
#include <regex.h>

typedef struct {
    int tiler;
    int renderer;
    int device;
} GPUUtilization;

typedef struct {
    int pid;
    const char *name;
    unsigned long long gpu_time_ns;
} GPUClient;

typedef struct {
    GPUUtilization util;
    GPUClient *clients;
    int client_count;
} GPUInfo;

static GPUUtilization getGPUUtilization() {
    GPUUtilization util = {0};

    io_service_t service = IOServiceGetMatchingService(
        kIOMainPortDefault,
        IOServiceMatching("AGXAcceleratorG15X")
    );
    if (service == IO_OBJECT_NULL) {
        // Try the older name
        service = IOServiceGetMatchingService(
            kIOMainPortDefault,
            IOServiceMatching("AGXAccelerator")
        );
    }
    if (service == IO_OBJECT_NULL) {
        return util;
    }

    CFMutableDictionaryRef props = NULL;
    kern_return_t kr = IORegistryEntryCreateCFProperties(
        service, &props, kCFAllocatorDefault, 0
    );
    IOObjectRelease(service);

    if (kr != KERN_SUCCESS || props == NULL) {
        return util;
    }

    // Utilization lives inside PerformanceStatistics sub-dictionary
    CFDictionaryRef perfStats = CFDictionaryGetValue(props, CFSTR("PerformanceStatistics"));
    CFDictionaryRef source = perfStats ? perfStats : (CFDictionaryRef)props;

    CFNumberRef tilerRef = CFDictionaryGetValue(source, CFSTR("Tiler Utilization %"));
    CFNumberRef rendererRef = CFDictionaryGetValue(source, CFSTR("Renderer Utilization %"));
    CFNumberRef deviceRef = CFDictionaryGetValue(source, CFSTR("Device Utilization %"));

    int val;
    if (tilerRef && CFNumberGetValue(tilerRef, kCFNumberIntType, &val)) {
        util.tiler = val;
    }
    if (rendererRef && CFNumberGetValue(rendererRef, kCFNumberIntType, &val)) {
        util.renderer = val;
    }
    if (deviceRef && CFNumberGetValue(deviceRef, kCFNumberIntType, &val)) {
        util.device = val;
    }

    CFRelease(props);
    return util;
}

static GPUInfo getGPUClients() {
    GPUInfo info = {0};

    // Walk children of accelerator — IOServiceGetMatchingServices doesn't
    // enumerate UserClient entries, but the child iterator does.
    io_service_t accel = IOServiceGetMatchingService(
        kIOMainPortDefault,
        IOServiceMatching("AGXAcceleratorG15X")
    );
    if (accel == IO_OBJECT_NULL) {
        accel = IOServiceGetMatchingService(
            kIOMainPortDefault,
            IOServiceMatching("AGXAccelerator")
        );
    }
    if (accel == IO_OBJECT_NULL) {
        return info;
    }

    io_iterator_t iter;
    kern_return_t kr = IORegistryEntryGetChildIterator(accel, kIOServicePlane, &iter);
    IOObjectRelease(accel);
    if (kr != KERN_SUCCESS) {
        return info;
    }

    // Temporary storage — allocate for up to 128 clients
    int capacity = 128;
    GPUClient *temp = (GPUClient *)calloc(capacity, sizeof(GPUClient));
    int count = 0;

    // pid -> index mapping for accumulation
    int pidMap[128];
    memset(pidMap, -1, sizeof(pidMap));

    io_service_t entry;
    while ((entry = IOIteratorNext(iter)) != IO_OBJECT_NULL) {
        CFMutableDictionaryRef props = NULL;
        kr = IORegistryEntryCreateCFProperties(entry, &props, kCFAllocatorDefault, 0);
        IOObjectRelease(entry);

        if (kr != KERN_SUCCESS || props == NULL) continue;

        // Parse IOUserClientCreator: "pid NNN, processname"
        CFStringRef creator = CFDictionaryGetValue(props, CFSTR("IOUserClientCreator"));
        if (creator == NULL) {
            CFRelease(props);
            continue;
        }

        char creatorBuf[256];
        if (!CFStringGetCString(creator, creatorBuf, sizeof(creatorBuf), kCFStringEncodingUTF8)) {
            CFRelease(props);
            continue;
        }

        int pid = 0;
        char procName[128] = {0};
        if (sscanf(creatorBuf, "pid %d, %127[^\n]", &pid, procName) < 2) {
            CFRelease(props);
            continue;
        }

        // Sum accumulatedGPUTime from AppUsage array
        unsigned long long totalTime = 0;
        CFArrayRef appUsage = CFDictionaryGetValue(props, CFSTR("AppUsage"));
        if (appUsage != NULL) {
            CFIndex usageCount = CFArrayGetCount(appUsage);
            for (CFIndex i = 0; i < usageCount; i++) {
                CFDictionaryRef usage = CFArrayGetValueAtIndex(appUsage, i);
                if (usage == NULL) continue;
                CFNumberRef gpuTime = CFDictionaryGetValue(usage, CFSTR("accumulatedGPUTime"));
                if (gpuTime != NULL) {
                    long long t = 0;
                    CFNumberGetValue(gpuTime, kCFNumberLongLongType, &t);
                    totalTime += (unsigned long long)t;
                }
            }
        }

        CFRelease(props);

        if (totalTime == 0) continue;

        // Find or create entry for this pid
        int idx = -1;
        for (int i = 0; i < count; i++) {
            if (temp[i].pid == pid) {
                idx = i;
                break;
            }
        }

        if (idx >= 0) {
            temp[idx].gpu_time_ns += totalTime;
        } else if (count < capacity) {
            temp[count].pid = pid;
            temp[count].name = strdup(procName);
            temp[count].gpu_time_ns = totalTime;
            count++;
        }
    }

    IOObjectRelease(iter);

    info.clients = temp;
    info.client_count = count;
    return info;
}

GPUInfo pollGPUInfo() {
    GPUInfo info = getGPUClients();
    info.util = getGPUUtilization();
    return info;
}

void freeGPUInfo(GPUInfo info) {
    if (info.clients != NULL) {
        for (int i = 0; i < info.client_count; i++) {
            if (info.clients[i].name != NULL) {
                free((void *)info.clients[i].name);
            }
        }
        free(info.clients);
    }
}

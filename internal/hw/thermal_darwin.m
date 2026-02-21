//go:build darwin

#import <Foundation/Foundation.h>

int getThermalStateValue() {
    NSProcessInfoThermalState state = [[NSProcessInfo processInfo] thermalState];
    return (int)state;
}

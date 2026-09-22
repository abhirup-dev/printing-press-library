//go:build darwin && cgo

package config

/*
#cgo darwin LDFLAGS: -framework Security -framework CoreFoundation
#include <CoreFoundation/CoreFoundation.h>
#include <Security/Security.h>
#include <stdlib.h>
#include <string.h>

static CFStringRef pp_string(const char *value) {
	return CFStringCreateWithCString(kCFAllocatorDefault, value, kCFStringEncodingUTF8);
}

static CFMutableDictionaryRef pp_query(const char *service, const char *account) {
	CFMutableDictionaryRef query = CFDictionaryCreateMutable(kCFAllocatorDefault, 0, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
	CFStringRef serviceRef = pp_string(service);
	CFStringRef accountRef = pp_string(account);
	CFDictionarySetValue(query, kSecClass, kSecClassGenericPassword);
	CFDictionarySetValue(query, kSecAttrService, serviceRef);
	CFDictionarySetValue(query, kSecAttrAccount, accountRef);
	CFRelease(serviceRef);
	CFRelease(accountRef);
	return query;
}

static int pp_keychain_read(const char *service, const char *account, void **outBytes, size_t *outLen) {
	*outBytes = NULL;
	*outLen = 0;
	CFMutableDictionaryRef query = pp_query(service, account);
	CFDictionarySetValue(query, kSecReturnData, kCFBooleanTrue);
	CFTypeRef result = NULL;
	OSStatus status = SecItemCopyMatching(query, &result);
	CFRelease(query);
	if (status != errSecSuccess) return (int)status;
	if (!result || CFGetTypeID(result) != CFDataGetTypeID()) {
		if (result) CFRelease(result);
		return (int)errSecDecode;
	}
	CFDataRef data = (CFDataRef)result;
	CFIndex length = CFDataGetLength(data);
	void *bytes = malloc((size_t)length);
	if (!bytes) { CFRelease(result); return (int)errSecAllocate; }
	memcpy(bytes, CFDataGetBytePtr(data), (size_t)length);
	*outBytes = bytes;
	*outLen = (size_t)length;
	CFRelease(result);
	return 0;
}

static int pp_keychain_write(const char *service, const char *account, const void *bytes, size_t length) {
	CFMutableDictionaryRef query = pp_query(service, account);
	CFDataRef data = CFDataCreate(kCFAllocatorDefault, (const UInt8 *)bytes, (CFIndex)length);
	if (!data) { CFRelease(query); return (int)errSecAllocate; }
	CFDictionarySetValue(query, kSecValueData, data);
	OSStatus status = SecItemAdd(query, NULL);
	CFRelease(data);
	CFRelease(query);
	if (status != errSecDuplicateItem) return (int)status;

	CFMutableDictionaryRef updateQuery = pp_query(service, account);
	CFMutableDictionaryRef update = CFDictionaryCreateMutable(kCFAllocatorDefault, 1, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
	CFDataRef updateData = CFDataCreate(kCFAllocatorDefault, (const UInt8 *)bytes, (CFIndex)length);
	CFDictionarySetValue(update, kSecValueData, updateData);
	status = SecItemUpdate(updateQuery, update);
	CFRelease(updateData);
	CFRelease(update);
	CFRelease(updateQuery);
	return (int)status;
}

static int pp_keychain_delete(const char *service, const char *account) {
	CFMutableDictionaryRef query = pp_query(service, account);
	OSStatus status = SecItemDelete(query);
	CFRelease(query);
	if (status == errSecItemNotFound) return 0;
	return (int)status;
}
*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"
)

func keychainRead() (string, error) {
	service := C.CString(keychainService)
	account := C.CString(keychainAccount)
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))
	var bytes unsafe.Pointer
	var length C.size_t
	status := C.pp_keychain_read(service, account, &bytes, &length)
	if status != 0 {
		return "", fmt.Errorf("keychain read status %d", int(status))
	}
	defer C.free(bytes)
	return strings.TrimSpace(string(C.GoBytes(bytes, C.int(length)))), nil
}

func keychainWrite(value string) error {
	service := C.CString(keychainService)
	account := C.CString(keychainAccount)
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))
	data := []byte(value)
	var ptr unsafe.Pointer
	if len(data) > 0 {
		ptr = unsafe.Pointer(&data[0])
	}
	status := C.pp_keychain_write(service, account, ptr, C.size_t(len(data)))
	if status != 0 {
		return fmt.Errorf("keychain write status %d", int(status))
	}
	return nil
}

func keychainDelete() error {
	service := C.CString(keychainService)
	account := C.CString(keychainAccount)
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))
	status := C.pp_keychain_delete(service, account)
	if status != 0 {
		return fmt.Errorf("keychain delete status %d", int(status))
	}
	return nil
}

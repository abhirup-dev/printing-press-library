//go:build darwin

package auth

/*
#cgo darwin LDFLAGS: -framework Security -framework CoreFoundation
#include <CoreFoundation/CoreFoundation.h>
#include <Security/Security.h>
#include <stdlib.h>
#include <string.h>

static CFStringRef tt_string(const char *value) {
	return CFStringCreateWithCString(kCFAllocatorDefault, value, kCFStringEncodingUTF8);
}

static void tt_query(CFMutableDictionaryRef query, const char *service, const char *account) {
	CFStringRef serviceRef = tt_string(service);
	CFStringRef accountRef = tt_string(account);
	CFDictionarySetValue(query, kSecClass, kSecClassGenericPassword);
	CFDictionarySetValue(query, kSecAttrService, serviceRef);
	CFDictionarySetValue(query, kSecAttrAccount, accountRef);
	CFRelease(serviceRef);
	CFRelease(accountRef);
}

static int tt_keychain_copy(const char *service, const char *account, void **out, size_t *outLen) {
	CFMutableDictionaryRef query = CFDictionaryCreateMutable(kCFAllocatorDefault, 0, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
	tt_query(query, service, account);
	CFDictionarySetValue(query, kSecReturnData, kCFBooleanTrue);
	CFDictionarySetValue(query, kSecMatchLimit, kSecMatchLimitOne);
	CFTypeRef result = NULL;
	OSStatus status = SecItemCopyMatching(query, &result);
	CFRelease(query);
	if (status != errSecSuccess) return (int)status;
	if (result == NULL || CFGetTypeID(result) != CFDataGetTypeID()) {
		if (result != NULL) CFRelease(result);
		return (int)errSecInvalidItemRef;
	}
	CFDataRef data = (CFDataRef)result;
	CFIndex length = CFDataGetLength(data);
	UInt8 *copy = (UInt8 *)malloc((size_t)length);
	if (copy == NULL) {
		CFRelease(result);
		return (int)errSecAllocate;
	}
	memcpy(copy, CFDataGetBytePtr(data), (size_t)length);
	*out = copy;
	*outLen = (size_t)length;
	CFRelease(result);
	return (int)errSecSuccess;
}

static int tt_keychain_write(const char *service, const char *account, const void *value, size_t valueLen) {
	CFMutableDictionaryRef query = CFDictionaryCreateMutable(kCFAllocatorDefault, 0, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
	tt_query(query, service, account);
	CFDataRef data = CFDataCreate(kCFAllocatorDefault, (const UInt8 *)value, (CFIndex)valueLen);
	CFDictionarySetValue(query, kSecValueData, data);
	OSStatus status = SecItemAdd(query, NULL);
	if (status == errSecDuplicateItem) {
		CFMutableDictionaryRef attrs = CFDictionaryCreateMutable(kCFAllocatorDefault, 0, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
		CFDictionarySetValue(attrs, kSecValueData, data);
		status = SecItemUpdate(query, attrs);
		CFRelease(attrs);
	}
	CFRelease(data);
	CFRelease(query);
	return (int)status;
}

static int tt_keychain_delete(const char *service, const char *account) {
	CFMutableDictionaryRef query = CFDictionaryCreateMutable(kCFAllocatorDefault, 0, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
	tt_query(query, service, account);
	OSStatus status = SecItemDelete(query);
	CFRelease(query);
	return (int)status;
}

static void tt_keychain_free(void *value) { free(value); }
*/
import "C"

import (
	"bytes"
	"fmt"
	"os/exec"
	"unsafe"
)

const keychainNotFound = C.errSecItemNotFound

func platformKeychainRead() ([]byte, error) {
	// Reads use Apple's security(1) reader rather than prompting a newly
	// promoted executable for an ACL decision. The secret is returned only in
	// memory; service/account are non-secret argv values and the password is
	// never an argv value.
	out, err := exec.Command("/usr/bin/security", "find-generic-password", "-a", Account, "-s", Service, "-w").Output()
	if err != nil {
		return nil, ErrNotFound
	}
	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return nil, ErrNotFound
	}
	return out, nil
}

func platformKeychainWrite(value []byte) error {
	service := C.CString(Service)
	account := C.CString(Account)
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))
	status := C.tt_keychain_write(service, account, unsafe.Pointer(&value[0]), C.size_t(len(value)))
	if status != C.errSecSuccess {
		return fmt.Errorf("Keychain write failed (OSStatus %d)", int(status))
	}
	return nil
}

func platformKeychainDelete() error {
	if err := exec.Command("/usr/bin/security", "delete-generic-password", "-a", Account, "-s", Service).Run(); err != nil {
		// Logout is idempotent; security(1) reports a missing item as a
		// non-zero exit, which is safe to treat as already cleared.
		return nil
	}
	return nil
}

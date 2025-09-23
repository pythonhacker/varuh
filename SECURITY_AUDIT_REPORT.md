# Varuh Security Audit Report

## Executive Summary

This document outlines security vulnerabilities and design flaws identified in the Varuh password manager during a comprehensive security audit. The analysis covers cryptographic implementation, database security, input validation, file handling, and overall system architecture.

## Critical Security Issues

### 1. SQL Injection Vulnerabilities

**Severity: HIGH**

**Location:** `db.go:526-530` and `db.go:702`

**Issue:** Direct string interpolation in SQL queries without proper parameterization.

```go
// Vulnerable code in searchDatabaseEntry function
searchTerm = fmt.Sprintf("%%%s%%", term)
query := db.Where(fmt.Sprintf("title like \"%s\"", searchTerm))

for _, field := range []string{"user", "url", "notes", "tags"} {
    query = query.Or(fmt.Sprintf("%s like \"%s\"", field, searchTerm))
}

// Vulnerable code in iterateEntries function  
rows, err = db.Model(&Entry{}).Order(fmt.Sprintf("%s %s", orderKey, order)).Rows()
```

**Impact:** Attackers could potentially execute arbitrary SQL commands by crafting malicious search terms or order parameters.

**Recommendation:** Use GORM's parameterized queries:
```go
query := db.Where("title LIKE ?", "%"+term+"%")
query = query.Or("user LIKE ?", "%"+term+"%")
// etc.
```

### 2. Weak Random Number Generation

**Severity: MEDIUM-HIGH**

**Location:** `crypto.go:478`

**Issue:** Use of `math/rand` with time-based seeding for password generation.

```go
rand.Seed(time.Now().UnixNano())
length = rand.Intn(4) + 12
```

**Impact:** Predictable random numbers could lead to weaker password generation.

**Recommendation:** Use `crypto/rand` consistently throughout the application.

### 3. Insecure File Permissions

**Severity: MEDIUM**

**Location:** Multiple files

**Issue:** Inconsistent file permission handling.

```go
// In utils.go - config files created with 0644
fh, err := os.OpenFile(configFile, os.O_RDWR, 0644)

// In crypto.go - encrypted files use 0600 (correct)
err = os.WriteFile(encDbPath, encText, 0600)
```

**Impact:** Configuration files may be readable by other users on the system.

**Recommendation:** Use 0600 permissions for all sensitive files.

## Design Flaws

### 4. Password Storage in Memory

**Severity: MEDIUM**

**Location:** Throughout the application

**Issue:** Passwords are stored in plain text in memory during operations and passed between functions as strings.

**Impact:** Passwords may persist in memory longer than necessary and could be exposed through memory dumps.

**Recommendation:** 
- Use secure memory clearing functions
- Minimize password lifetime in memory
- Consider using byte slices that can be zeroed

### 5. Insufficient Input Validation

**Severity: MEDIUM**

**Location:** Multiple locations

**Issues:**
- No length limits on user inputs (titles, URLs, notes)
- No validation of URL format beyond basic HTTP/HTTPS prefix
- No sanitization of custom field names

**Impact:** Potential for denial of service or data corruption.

**Recommendation:** Implement comprehensive input validation and sanitization.

### 6. Signal Handling Race Conditions

**Severity: LOW-MEDIUM**

**Location:** `actions.go:40-50` and `actions.go:82-92`

**Issue:** Signal handlers may not properly clean up resources or may cause race conditions.

```go
go func() {
    sig := <-sigChan
    fmt.Println("Received signal", sig)
    // Reencrypt
    encryptDatabase(defaultDB, &encPasswd)
    os.Exit(1)
}()
```

**Impact:** Potential data loss or corruption during unexpected termination.

**Recommendation:** Implement proper cleanup mechanisms and avoid race conditions.

### 7. Clipboard Security

**Severity: LOW-MEDIUM**

**Location:** `utils.go:599`

**Issue:** Passwords are copied to system clipboard without automatic clearing.

```go
func copyPasswordToClipboard(passwd string) {
    clipboard.WriteAll(passwd)
}
```

**Impact:** Passwords may remain in clipboard history accessible to other applications.

**Recommendation:** Implement clipboard clearing after a timeout or provide user notification.

## Cryptographic Concerns

### 8. Argon2 Parameters

**Severity: LOW**

**Location:** `crypto.go:68`

**Issue:** Argon2 parameters may be insufficient for current security standards.

```go
key = argon2.Key([]byte(passPhrase), salt, 3, 32*1024, 4, KEY_SIZE)
```

**Impact:** Weaker key derivation than recommended.

**Recommendation:** Use Argon2id with higher memory (64MB+) and iterations (4+).

### 9. HMAC Implementation

**Severity: LOW**

**Location:** `crypto.go:183-186`

**Issue:** HMAC-SHA512 is used for authentication, which is acceptable but SHA-256 would be sufficient.

**Impact:** Minor performance impact, no security impact.

**Recommendation:** Consider using HMAC-SHA256 for better performance.

## File System Security

### 10. Temporary File Handling

**Severity: MEDIUM**

**Location:** `export.go:171` and `export.go:188`

**Issue:** Temporary files may not be properly cleaned up in all error scenarios.

```go
tmpFile = randomFileName(os.TempDir(), ".tmp")
// ... operations ...
os.Remove(tmpFile) // May not execute in error cases
```

**Impact:** Sensitive data may remain in temporary files.

**Recommendation:** Use defer statements for cleanup and ensure proper error handling.

### 11. File Overwrite Operations

**Severity: MEDIUM**

**Location:** `crypto.go:194-203`

**Issue:** Atomic file operations are attempted but may fail partially.

```go
err = os.WriteFile(encDbPath, encText, 0600)
if err == nil {
    err = os.WriteFile(dbPath, encText, 0600)
    if err == nil {
        os.Remove(encDbPath)
    }
}
```

**Impact:** Potential data loss if operations fail partially.

**Recommendation:** Implement proper atomic file operations using rename operations.

## Database Security

### 12. Database Schema Exposure

**Severity: LOW**

**Location:** `db.go:18-60`

**Issue:** Database schema is well-documented and predictable.

**Impact:** Makes it easier for attackers to understand the data structure.

**Recommendation:** Consider obfuscating field names or using generic field names.

### 13. No Database Encryption at Rest

**Severity: MEDIUM**

**Location:** Database operations

**Issue:** SQLite databases are stored unencrypted when not using the application's encryption.

**Impact:** Database files may be readable if accessed directly.

**Recommendation:** Consider using SQLCipher or similar encrypted database solutions.

## Recommendations Summary

### Immediate Actions Required:
1. **Fix SQL injection vulnerabilities** - Use parameterized queries
2. **Implement proper input validation** - Add length limits and sanitization
3. **Fix file permissions** - Use 0600 for all sensitive files
4. **Improve random number generation** - Use crypto/rand consistently

### Medium Priority:
1. **Implement secure memory handling** - Clear sensitive data from memory
2. **Improve file operations** - Use atomic operations
3. **Add clipboard security** - Implement timeout or clearing
4. **Enhance signal handling** - Prevent race conditions

### Long-term Improvements:
1. **Upgrade cryptographic parameters** - Use stronger Argon2 settings
2. **Consider database encryption** - Use SQLCipher
3. **Implement audit logging** - Track security-relevant events
4. **Add comprehensive testing** - Security-focused test suite

## Conclusion

While Varuh implements good cryptographic practices with AES-256 and XChaCha20-Poly1305, several critical security vulnerabilities need immediate attention. The SQL injection vulnerabilities are the most serious concern and should be addressed immediately. The overall architecture is sound, but implementation details need refinement to meet security best practices.

The application would benefit from a security-focused refactoring to address these issues systematically, with particular attention to input validation, memory management, and file handling security.
# Adding Language Support to Neuron

This guide walks you through adding support for a new programming language to the Neuron code execution platform.

## 📋 Overview

Adding a new language requires changes in **4 key areas**:

1. **Language Registry** - Define language configuration
2. **Validator** - Add security validation
3. **Docker Pool** - Configure container pool
4. **Error Detection** - Add language-specific error parsing

---

## 🔧 Step-by-Step Guide

### Step 1: Add Language Configuration

**File**: `internal/registry/language.go`

Add your language to the `LanguageRegistry` map:

```go
var LanguageRegistry = map[string]LanguageConfig{
    // ... existing languages ...
    
    "rust": {
        Name:        "rust",
        Validator:   ValidateAndSanitizeRust,  // We'll create this next
        DockerImage: "rust:1.75-alpine",
        BaseName:    "main",
        Ext:         "rs",
        RunCmd: func(n FileNames) string {
            return fmt.Sprintf(
                "rustc %s -o %s && ./%s < input.txt",
                n.FullName, n.BaseName, n.BaseName,
            )
        },
        CreditCost: 5,
    },
}
```

#### Configuration Fields:

| Field | Description | Example |
|-------|-------------|---------|
| `Name` | Language identifier (lowercase) | `"rust"` |
| `Validator` | Function to validate/sanitize code | `ValidateAndSanitizeRust` |
| `DockerImage` | Docker image to use | `"rust:1.75-alpine"` |
| `BaseName` | Base filename (without extension) | `"main"` or `"Main"` |
| `Ext` | File extension | `"rs"` |
| `RunCmd` | Function that returns shell command to compile/run | See example above |
| `CreditCost` | Credits to charge per execution | `5` |

---

### Step 2: Add Code Validator

**File**: `internal/registry/validators.go`

Create a validation function to prevent malicious code:

```go
func ValidateAndSanitizeRust(code string) error {
    // 1. Size limit
    if len(code) > 256*1024 {
        return fmt.Errorf("rust code too large (>256KB)")
    }

    // 2. Non-printable characters
    for _, r := range code {
        if !unicode.IsPrint(r) && r != '\n' && r != '\t' {
            return fmt.Errorf("contains invalid characters")
        }
    }

    // 3. Basic language heuristics
    if !strings.Contains(code, "fn main()") {
        return fmt.Errorf("missing main() function")
    }

    // 4. Dangerous keywords/modules
    blocked := []string{
        "std::process",
        "std::fs",
        "std::net",
        "std::os",
        "unsafe {",
        "Command::new",
        "File::create",
        "File::open",
    }
    
    for _, bad := range blocked {
        if strings.Contains(code, bad) {
            return fmt.Errorf("code contains forbidden keyword: %s", bad)
        }
    }
    
    return nil
}
```

#### Security Checklist:

- ✅ **Size limit** - Prevent resource exhaustion
- ✅ **Character validation** - Block non-printable characters
- ✅ **Language structure** - Ensure valid code structure
- ✅ **Dangerous APIs** - Block file I/O, network, process execution

---

### Step 3: Configure Docker Pool

**File**: `config/docker_pool.go`

Add pool configuration for your language:

```go
func DockerPools() []DockerPoolConfig {
    return []DockerPoolConfig{
        // ... existing pools ...
        
        {
            Language:       "rust",
            Image:          "rust:1.75-alpine",
            InitSize:       1,                    // Initial containers
            MaxSize:        2,                    // Maximum containers
            HealthCmd:      []string{"rustc", "--version"},
            HealthInterval: 20 * time.Second,
        },
    }
}
```

#### Pool Configuration:

| Field | Description | Recommended Value |
|-------|-------------|-------------------|
| `Language` | Must match registry name | Same as Step 1 |
| `Image` | Docker image (should match registry) | Same as Step 1 |
| `InitSize` | Containers created at startup | `1-2` |
| `MaxSize` | Maximum pool size | `2-5` |
| `HealthCmd` | Command to verify container health | `["rustc", "--version"]` |
| `HealthInterval` | How often to check health | `20s - 60s` |

---

### Step 4: Add Error Detection

**File**: `pkg/sandbox/docker/lang.go`

Add language-specific error detection in the `DetectError` function:

```go
func DetectError(language, stdout, stderr string) (models.SandboxError, string) {
    s := stderr
    c := stdout + "\n" + stderr
    
    // ... existing language checks ...
    
    // Rust
    if language == "rust" {
        // Compilation errors
        if strings.Contains(s, "error:") ||
           strings.Contains(s, "error[E") ||
           strings.Contains(s, "cannot find") {
            return models.ErrCompilationError, models.MsgCompilationError
        }
        
        // Runtime errors
        if strings.Contains(c, "thread 'main' panicked") ||
           strings.Contains(c, "stack backtrace:") {
            return models.ErrRuntimeError, models.MsgRuntimeError
        }
    }
    
    // ... rest of function ...
}
```

#### Error Types:

- `ErrCompilationError` - Syntax errors, missing symbols
- `ErrRuntimeError` - Panics, exceptions, crashes
- `ErrTLE` - Time limit exceeded (handled automatically)
- `ErrMLE` - Memory limit exceeded (handled automatically)

---

## 🧪 Testing Your Language

### Manual Testing

Start the services and submit a test job:

```bash
# Terminal 1: Start API
air -c .air.api.toml

# Terminal 2: Start Worker
air -c .air.worker.toml

# Terminal 3: Submit test job
curl -X POST http://localhost:8080/api/v1/runner/submit \
  -H "X-API-KEY: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "language": "rust",
    "code": "fn main() { println!(\"Hello!\"); }",
    "input": ""
  }'
```

**Verification Checklist:**
- ✅ Code validates successfully
- ✅ Container pool initializes
- ✅ Code executes and returns output
- ✅ Security validation blocks dangerous code

---

## 📝 Checklist

Before submitting your PR, ensure:

- [ ] Added language to `LanguageRegistry`
- [ ] Created validator function in `validators.go`
- [ ] Added pool configuration in `docker_pool.go`
- [ ] Added error detection in `lang.go`
- [ ] Tested with valid code
- [ ] Tested security blocking
- [ ] Verified container pool initialization
- [ ] Updated documentation

---

## 🎯 Example: Complete Rust Implementation

Here's a complete example for Rust:

### 1. Registry Entry
```go
"rust": {
    Name:        "rust",
    Validator:   ValidateAndSanitizeRust,
    DockerImage: "rust:1.75-alpine",
    BaseName:    "main",
    Ext:         "rs",
    RunCmd: func(n FileNames) string {
        return fmt.Sprintf("rustc %s -o %s && ./%s < input.txt", 
            n.FullName, n.BaseName, n.BaseName)
    },
    CreditCost: 5,
},
```

### 2. Validator
```go
func ValidateAndSanitizeRust(code string) error {
    if len(code) > 256*1024 {
        return fmt.Errorf("rust code too large")
    }
    if !strings.Contains(code, "fn main()") {
        return fmt.Errorf("missing main() function")
    }
    blocked := []string{"std::process", "std::fs", "std::net", "unsafe {"}
    for _, bad := range blocked {
        if strings.Contains(code, bad) {
            return fmt.Errorf("forbidden: %s", bad)
        }
    }
    return nil
}
```

### 3. Pool Config
```go
{
    Language:       "rust",
    Image:          "rust:1.75-alpine",
    InitSize:       1,
    MaxSize:        2,
    HealthCmd:      []string{"rustc", "--version"},
    HealthInterval: 20 * time.Second,
},
```

### 4. Error Detection
```go
if language == "rust" {
    if strings.Contains(s, "error:") {
        return models.ErrCompilationError, models.MsgCompilationError
    }
    if strings.Contains(c, "panicked") {
        return models.ErrRuntimeError, models.MsgRuntimeError
    }
}
```

---

### Optimizing Compilation Time

For compiled languages, consider:
- Using smaller base images (alpine)
- Pre-installing common dependencies
- Caching compiled artifacts (if applicable)

### Resource Limits

Container resource limits are defined in `pool_manager.go`:
- CPU: Unlimited (relies on Docker host limits)
- Memory: Enforced via cgroups
- Network: Disabled (`NetworkMode: "none"`)
- Filesystem: Read-only root + writable /tmp

---

## 🤝 Need Help?

- Open an issue on GitHub
- Check existing language implementations
- Review the [CONTRIBUTING.md](./CONTRIBUTING.md) guide

---

**Happy Language Adding! 🎉**

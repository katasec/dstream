# HCL Locals Implementation Checklist

**Goal**: Add HCL `locals` support to solve table duplication problem in DStream configurations.

**Time Estimate**: 2-3 hours total implementation

## Problem Statement

Current configuration requires table duplication:
```hcl
# ❌ CURRENT: Duplication risk
task "mssql-cdc-to-asb" {
  input {
    config {
      tables = ["Orders", "Customers", "Products"]  # List 1
    }
  }
  output {
    config {
      tables = ["Orders", "Customers", "Products"]  # List 2 - can drift!
    }
  }
}
```

**Solution**: HCL `locals` with single source of truth:
```hcl
# ✅ SOLUTION: Single source of truth
locals {
  tables = ["Orders", "Customers", "Products"]
  env_name = "production"
}

task "mssql-cdc-to-asb" {
  input {
    config {
      tables = local.tables  # Single reference
    }
  }
  output {
    config {
      tables = local.tables  # Same reference - no duplication!
    }
  }
}
```

## 📋 Simplified HCL Locals Implementation Checklist

### 🔍 **Phase 1: Analysis & Discovery (30-45 minutes)**

**Current Architecture Investigation:**
- [ ] **Find HCL parsing location**: Locate where `dstream.hcl` is currently parsed
- [ ] **Template processing order**: Determine if `{{ env "VAR" }}` happens before/after HCL parsing
- [ ] **Config struct mapping**: Find Go structs that represent parsed HCL configuration
- [ ] **Error handling patterns**: See how current HCL parsing errors are displayed
- [ ] **Validation flow**: Understand existing config validation and user feedback

**Template Integration Analysis:**
- [ ] **Template engine location**: Find where `{{ env "VAR" }}` is processed
- [ ] **Interpolation timing**: When do templates get resolved in the pipeline?
- [ ] **Variable scope**: How do templates currently access values?
- [ ] **Error handling**: How template errors are reported to users

### 🏗️ **Phase 2: Simplified HCL Grammar Design**

**Locals Block Specification (Simple Key-Value Only):**
```hcl
# Optional block - must validate if present
locals {
  tables = ["Orders", "Customers", "Products"]
  env_name = "production"
  database_name = "OrdersDB"
  poll_interval = "5s"
}

# Usage in tasks - simple references only
task "example" {
  input {
    config {
      tables = local.tables           # Array reference
      environment = local.env_name    # String reference  
      database = local.database_name  # String reference
      poll_interval = local.poll_interval  # String reference
    }
  }
}
```

**Simplified Validation Requirements:**
- [ ] **Optional locals**: If no `locals` block, continue normally
- [ ] **Single locals block**: Error if multiple `locals` blocks present  
- [ ] **Simple attributes only**: Only `key = value`, no nested blocks or objects
- [ ] **Valid references**: Error if `local.undefined_var` is used
- [ ] **Type compatibility**: Ensure `local.tables` matches expected config types

### 🔧 **Phase 3: Integration with Template System**

**Template Processing Order:**
```
1. Load HCL file content
2. Process {{ env "VAR" }} templates → String substitution  
3. Parse HCL with locals evaluation → Structured data
4. Validate config against Go structs
```

**Simplified Template + Locals Interaction:**
```hcl
locals {
  env_prefix = "{{ env \"ENVIRONMENT\" }}"    # Templates work in locals
  database_name = "{{ env \"DB_NAME\" }}"     # Templates work in locals
}

task "example" {
  input {
    config {
      database = local.database_name          # Simple reference
    }
  }
}
```

**Requirements:**
- [ ] **Template-in-locals**: `{{ env "VAR" }}` should work inside `locals` block values
- [ ] **Simple references only**: Only `local.var`, no `local.obj.prop` or `local.arr[0]`
- [ ] **Error precedence**: Template errors vs locals errors - which reports first?

### 🎯 **Phase 4: Simplified Error Handling**

**Error Categories & Messages:**
```bash
# Missing locals reference
❌ Error: Reference to undefined local value
  → local.undefined_table is not defined in locals block
  → Available locals: tables, env_name, database_name

# Invalid locals syntax (simplified)
❌ Error: Invalid locals block syntax
  → Line 3: locals block can only contain simple attribute assignments
  → Use: tables = ["Orders", "Customers"]
  → Not: nested { blocks = "not allowed" }

# Multiple locals blocks
❌ Error: Multiple locals blocks found
  → Only one locals block is allowed per configuration file
  → Found blocks at lines 5 and 23
```

**Simplified Validation Requirements:**
- [ ] **Clear error locations**: Show file, line, column for errors
- [ ] **Helpful suggestions**: Show available locals when reference fails
- [ ] **Simple syntax only**: Reject nested blocks in locals
- [ ] **Reference validation**: Check all `local.xxx` references are valid

### 🔍 **Phase 5: Grammar Linting & Validation**

**Separate Linting Command:**
```bash
dstream lint dstream.hcl           # Validate HCL file without running
```

**Simplified Linting Requirements:**
- [ ] **Grammar validation**: HCL syntax correctness
- [ ] **Simple locals validation**: Key-value pairs only, no nesting
- [ ] **Template validation**: `{{ env "VAR" }}` syntax in locals values
- [ ] **Config validation**: Final config matches expected Go structs
- [ ] **Performance**: Fast validation without heavy processing

### 🧪 **Phase 6: Simplified Testing Strategy (30-45 minutes)**

**Test Cases:**
- [ ] **No locals block**: Existing configs continue working
- [ ] **Simple string locals**: `local.env_name = "production"`
- [ ] **Simple array locals**: `local.tables = ["Orders", "Customers"]`
- [ ] **Templates in locals**: `local.db = "{{ env \"DB_NAME\" }}"`
- [ ] **Invalid references**: Clear error for `local.undefined`
- [ ] **Multiple locals**: Error handling for duplicate blocks
- [ ] **Nested blocks rejected**: Error for `locals { nested { } }`

**Integration Tests:**
- [ ] **Full pipeline**: `locals` → template processing → config validation → provider execution
- [ ] **Error propagation**: Locals errors surface properly to user
- [ ] **Backward compatibility**: All existing `dstream.hcl` files continue working

### 🔧 **Phase 7: Core Implementation (60-90 minutes)**

**Code Changes Required:**
- [ ] **HCL parsing enhancement**: Add simple `locals` block evaluation with `hcl.EvalContext`
- [ ] **Template integration**: Ensure locals work with existing `{{ env }}` system  
- [ ] **Simple validation**: Reject complex locals, accept only key-value pairs
- [ ] **Config structs**: Add locals validation to config loading
- [ ] **CLI command**: Add `dstream lint` for separate validation
- [ ] **Documentation**: Update examples with simple locals usage

**Go Dependencies:**
- [ ] **Verify hcl/v2**: Ensure `github.com/hashicorp/hcl/v2` is available
- [ ] **Template library**: Check current template engine (text/template?)
- [ ] **Error formatting**: Use existing error display patterns

**Core Implementation Example:**
```go
import (
    "github.com/hashicorp/hcl/v2"
    "github.com/hashicorp/hcl/v2/hclsyntax"
    "github.com/zclconf/go-cty/cty"
)

func parseHCLWithLocals(filename string) (*Config, error) {
    // 1. Parse HCL file
    file, diags := hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
    
    // 2. Create EvalContext with locals support
    ctx := &hcl.EvalContext{
        Variables: map[string]cty.Value{},
        Functions: map[string]function.Function{},
    }
    
    // 3. Evaluate locals block first
    if localsBlock := file.Body.(*hclsyntax.Body).Blocks.OfType("locals"); len(localsBlock) > 0 {
        localsAttrs, _ := localsBlock[0].Body.JustAttributes()
        for name, attr := range localsAttrs {
            val, _ := attr.Expr.Value(ctx)
            ctx.Variables[name] = val
        }
    }
    
    // 4. Evaluate task blocks with locals context
    var config Config
    diags = gohcl.DecodeBody(file.Body, ctx, &config)
    
    return &config, nil
}
```

### 🎯 **Phase 8: Simplified Edge Cases**

**Edge Cases to Handle:**
- [ ] **Empty locals block**: `locals {}` should be valid but do nothing
- [ ] **Reserved keywords**: Prevent `locals { env = "..." }` conflicting with templates
- [ ] **Simple expressions only**: Just `local.var`, no indexing or property access
- [ ] **Cross-task references**: Locals should be global per file

**Performance Considerations:**
- [ ] **Caching**: Don't re-evaluate locals for each task
- [ ] **Parse time**: Simple locals shouldn't slow config loading

### ✅ **Phase 9: Validation & Documentation**

**User Documentation:**
- [ ] **Simple HCL examples**: Show basic locals patterns for table duplication problem
- [ ] **Migration guide**: Convert duplicated config to simple locals
- [ ] **Error reference**: Document all possible simple locals errors

## ⏱️ Time Breakdown

### **Total Estimate: 2-3 hours**

- **Phase 1 (Discovery)**: 30-45 minutes
- **Phase 7 (Core Implementation)**: 60-90 minutes  
- **Phase 6 (Testing)**: 30-45 minutes

### **Confidence Level: High** ✅

**Why this timeline is realistic:**
- ✅ **Simple scope**: Just key-value locals, no complex features
- ✅ **Standard HCL pattern**: `hcl.EvalContext` is well-documented
- ✅ **Small change surface**: ~20-50 lines of Go code
- ✅ **Clear goal**: Solve table duplication problem specifically

**Potential blockers that could extend timeline:**
- 🤔 **Complex template integration**: If template/locals interaction is tricky (+30-60 min)
- 🤔 **Unexpected config structure**: If current HCL parsing is complex (+30 min)
- 🤔 **Testing edge cases**: If we discover more validation needed (+30 min)

## 🎯 **Key Decision Points:**

1. **Template Integration**: Do templates process before or after locals evaluation?
2. **Reference Syntax**: Only `local.var`, no complex expressions
3. **Error Precedence**: Which errors show first - template, locals, or config validation?
4. **Scope**: Locals are global per HCL file
5. **Syntax**: Only simple key-value pairs, no nested objects or blocks

## 🚀 **Success Criteria**

Once complete, this should work:
```hcl
locals {
  tables = ["Orders", "Customers", "Products"]
}

task "mssql-to-asb" {
  input {
    config {
      tables = local.tables
    }
  }
  output {
    config {
      tables = local.tables  # Same reference - no duplication!
    }
  }
}
```

```bash
# Full workflow should work:
dstream init mssql-to-asb    # Creates 3 queues automatically
dstream run mssql-to-asb     # Routes to correct queues based on table metadata
dstream destroy mssql-to-asb # Destroys all 3 queues
```

**This simplified approach focuses on solving the table duplication problem without over-engineering. We can extend later if needed.**

---

**File created**: `HCL_LOCALS_IMPLEMENTATION_CHECKLIST.md` in `/Users/writeameer/progs/dstream/dstream/`
**Ready to continue tomorrow**: All planning and analysis documented for efficient implementation.
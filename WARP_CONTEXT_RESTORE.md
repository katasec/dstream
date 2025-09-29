# WARP Context Restoration Prompt

**Use this prompt when starting a new Warp session to quickly restore full project context:**

---

## 🎯 **Recommended Context Restoration Prompt:**

```
Please read the WARP.md file to understand the DStream project and provide me with a summary of the current state and next steps.

Key context:
- I'm working on the DStream data streaming orchestration system
- We've moved from legacy gRPC plugins to modern stdin/stdout providers
- Project is organized under ~/progs/dstream/ with consolidated structure
- Each provider has self-documenting Makefiles for building single-file binaries
- Current working task: counter-to-console (modern provider orchestration)
- All repositories have been recently updated and pushed

Please confirm:
1. What stage the project is currently at
2. How to build and test the providers
3. What the current working architecture looks like
4. Any next logical development steps

Environment: macOS, PowerShell 7.5.2, .NET 9.0
```

---

## 📋 **Alternative Short Version:**

```
Please read ~/progs/dstream/WARP.md for full context on the DStream project. 

This is a data streaming system with modern stdin/stdout providers (not legacy gRPC plugins). 
I need to understand: current state, how to build/test providers, and next development steps.

Environment: macOS, PowerShell, .NET 9.0
```

---

## 🚀 **Usage Instructions:**

1. **Copy and paste** either prompt above into a new Warp session
2. **Wait for me to read** WARP.md and assess the project state
3. **I'll provide** a summary of current state and next steps
4. **Continue development** without spending time on context restoration

---

## 📝 **What This Achieves:**

- ✅ **Instant Context**: Full project understanding in one prompt
- ✅ **Current State Assessment**: Where development currently stands
- ✅ **Build Instructions**: How to work with providers
- ✅ **Next Steps**: Logical progression of development tasks
- ✅ **No Time Lost**: Skip hours of context rebuilding

---

**Last Updated**: September 15, 2025  
**Project Stage**: Modern provider architecture with self-documenting build system  
**Key Achievement**: Single-file binaries with stdin/stdout communication working
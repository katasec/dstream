# Generic Plugin Configuration Implementation

This task list tracks the implementation of strongly-typed generic configuration support for DStream plugins.

## Phase 1: Katasec.DStream.Plugin Library Updates

- [ ] **Task 1.1:** Create Generic Interface (`IDStreamPlugin<TConfig>`) in new file
  - Project: Katasec.DStream.Plugin
  - File: `/src/Katasec.DStream.Plugin/Interfaces/IDStreamPluginGeneric.cs`
  - Description: Create a new generic interface that extends the existing one
  - Implementation:
    ```csharp
    public interface IDStreamPlugin<TConfig> : IDStreamPlugin
    {
        Task ProcessAsync(IInput input, IOutput output, TConfig config, CancellationToken cancellationToken);
    }
    ```

- [ ] **Task 1.2:** Create Configuration Conversion Utilities
  - Project: Katasec.DStream.Plugin
  - File: `/src/Katasec.DStream.Plugin/ConfigurationUtils.cs`
  - Description: Create utility methods for converting between protobuf and typed configs
  - Implementation:
    - Create methods to extract values from Protobuf Value objects
    - Add support for common types (string, number, boolean, etc.)
    - Add debug logging in conversion methods to trace the conversion process

- [ ] **Task 1.3:** Update DStreamPluginHost for Generic Support
  - Project: Katasec.DStream.Plugin
  - File: `/src/Katasec.DStream.Plugin/DStreamPluginHost.cs`
  - Description: Update the host to handle both generic and non-generic plugins
  - Implementation:
    - Add detection of generic interface implementation
    - Add deserialization logic before calling plugin's ProcessAsync
    - Add detailed logging for configuration conversion

## Phase 2: Sample Plugin Implementation

- [x] **Task 2.1:** Create Configuration Classes for Plugins
  - Project: dstream-dotnet-test
  - Files: 
    - `/CounterPlugin.cs` (CounterPluginConfig)
    - `/GenericCounterPlugin.cs` (GenericCounterConfig)
  - Description: Create strongly typed configuration classes for plugins
  - ✅ Completed: Created configuration classes with appropriate properties

- [x] **Task 2.2:** Add Configuration Extraction Helpers
  - Project: dstream-dotnet-test
  - File: `/CounterPlugin.cs`
  - Description: Create helper methods to extract values from Protobuf Value objects
  - ✅ Completed: Added ExtractNumberValue and FromDictionary methods

- [ ] **Task 2.3:** Update GenericCounterPlugin for Generic Interface
  - Project: dstream-dotnet-test
  - File: `/GenericCounterPlugin.cs`
  - Description: Update the plugin to implement IDStreamPlugin<GenericCounterConfig>
  - Dependencies: Requires Task 1.1 completion
  - Implementation:
    - Change interface to IDStreamPlugin<GenericCounterConfig>
    - Update ProcessAsync method signature
    - Remove manual deserialization code

## Phase 3: Integration and Testing

- [ ] **Task 3.1:** Update Program.cs for Generic Interface
  - Project: dstream-dotnet-test
  - File: `/Program.cs`
  - Description: Update the program to use the generic plugin host once available
  - Dependencies: Requires Task 1.3 completion
  - Implementation:
    - Update DStreamPluginHost instantiation to use generic version
    - Ensure proper registration of providers

- [ ] **Task 3.2:** Update GenericProgram.cs
  - Project: dstream-dotnet-test
  - File: `/GenericProgram.cs`
  - Description: Update the program to use the generic plugin host
  - Dependencies: Requires Task 1.3 completion
  - Implementation:
    - Convert to class with Main method to avoid top-level statement conflicts
    - Update DStreamPluginHost instantiation

- [ ] **Task 3.3:** Create Test HCL Configuration
  - Project: dstream
  - File: `/examples/generic-counter-test.hcl`
  - Description: Create a test configuration for the generic counter plugin
  - Dependencies: None (can be done independently)
  - Implementation:
    - Add interval, prefix, and includeTimestamp configuration
    - Configure input and output providers

## Phase 4: Testing and Validation

- [ ] **Task 4.1:** Test CounterPlugin with Current Implementation
  - Description: Test the current implementation with manual configuration conversion
  - Dependencies: None (can be done now)
  - Steps:
    1. Build the plugin with `./build.ps1 publish`
    2. Run it with `dstream run dotnet-counter`
    3. Verify the logs show proper configuration extraction

- [ ] **Task 4.2:** Test GenericCounterPlugin with Generic Interface
  - Description: Test the generic implementation after framework updates
  - Dependencies: Requires Tasks 1.1-1.3, 2.3, and 3.1-3.3 completion
  - Steps:
    1. Build the plugin with `./build.ps1 publish`
    2. Run it with `dstream run generic-counter`
    3. Verify the logs show proper configuration deserialization

## Phase 5: Framework Integration

- [ ] **Task 5.1:** Move Configuration Utilities to Framework
  - Project: Katasec.DStream.Plugin
  - File: `/src/Katasec.DStream.Plugin/ConfigurationUtils.cs`
  - Description: Move the configuration extraction helpers to the framework
  - Dependencies: Requires Task 4.1 completion (validation of current approach)
  - Implementation:
    - Extract ExtractNumberValue and similar methods from CounterPlugin
    - Create a comprehensive configuration utility class
    - Add support for additional types (arrays, nested objects)

- [ ] **Task 5.2:** Update CounterPlugin to Use Framework Utilities
  - Project: dstream-dotnet-test
  - File: `/CounterPlugin.cs`
  - Description: Update the plugin to use the framework utilities
  - Dependencies: Requires Task 5.1 completion
  - Implementation:
    - Replace local extraction methods with framework utilities
    - Update configuration handling to use the new utilities

## Phase 6: Documentation and Examples

- [ ] **Task 6.1:** Document Configuration Approach
  - Project: Katasec.DStream.Plugin
  - File: `/docs/configuration.md`
  - Description: Document the configuration approach and best practices
  - Dependencies: Requires all previous phases to be completed
  - Implementation:
    - Explain the generic interface and its benefits
    - Provide examples of configuration classes
    - Document the conversion utilities
    - Add migration guide for existing plugins

- [ ] **Task 6.2:** Create Example Plugins
  - Project: dstream-dotnet-test
  - Files: `/examples/`
  - Description: Create example plugins showing different configuration scenarios
  - Dependencies: Requires Phase 5 completion
  - Implementation:
    - Simple configuration example
    - Nested configuration example
    - Array/list configuration example
    - Complex configuration with multiple types

## Notes

- This implementation plan focuses on a gradual approach to minimize disruption
- We start with framework changes, then update plugins to use the new features
- Each phase can be tested independently before moving to the next
- The approach maintains backward compatibility throughout the transition
- Testing is integrated at multiple points to validate the approach

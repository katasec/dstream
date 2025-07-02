# Generic Plugin Configuration Implementation

This task list tracks the implementation of strongly-typed generic configuration support for DStream plugins.

## Phase 1: Katasec.DStream.Plugin Library Updates

- [ ] **Task 1.1:** Create Generic Interface (`IDStreamPlugin<TConfig>`) in new file
  - File: `/src/Katasec.DStream.Plugin/Interfaces/IDStreamPluginGeneric.cs`
  - Description: Create a new generic interface that doesn't replace the existing one yet

- [ ] **Task 1.2:** Create Configuration Conversion Utilities
  - File: `/src/Katasec.DStream.Plugin/ConfigurationUtils.cs`
  - Description: Create utility methods for converting between protobuf and typed configs
  - Add debug logging in conversion methods to trace the conversion process

- [ ] **Task 1.3:** Create Base Plugin Class
  - File: `/src/Katasec.DStream.Plugin/DStreamPluginBase.cs`
  - Description: Create a base class that implements both interfaces for backward compatibility
  - Add debug logging for the bridge method

- [ ] **Task 1.4:** Update DStreamPluginHost for Generic Support
  - File: `/src/Katasec.DStream.Plugin/DStreamPluginHostGeneric.cs`
  - Description: Create a new generic host class that doesn't replace the existing one yet
  - Add detailed logging for configuration conversion and plugin initialization

## Phase 2: Sample Plugin Implementation

- [ ] **Task 2.1:** Create a Sample Generic Plugin
  - Project: dstream-dotnet-test
  - File: `/GenericCounterPlugin.cs`
  - Description: Create a new plugin that uses the generic interface
  - Add logging to show the typed configuration values being used

- [ ] **Task 2.2:** Create a Program Entry Point for the Generic Plugin
  - Project: dstream-dotnet-test
  - File: `/GenericProgram.cs`
  - Description: Create a new program entry point for the generic plugin

## Phase 3: Integration and Testing

- [ ] **Task 3.1:** Create Test HCL Configuration
  - Project: dstream
  - File: `/examples/generic-plugin-test.hcl`
  - Description: Create a test configuration for the generic plugin

- [ ] **Task 3.2:** Test the Generic Plugin
  - Steps:
    1. Build the generic plugin
    2. Run it with the test configuration
    3. Verify the logs show proper configuration conversion and usage
  - Logging to verify:
    - Log the raw protobuf configuration received
    - Log the converted typed configuration
    - Log the usage of typed configuration values during execution

## Phase 4: Backward Compatibility Bridge

- [ ] **Task 4.1:** Update IDStreamPlugin Interface
  - Project: Katasec.DStream.Plugin
  - File: `/src/Katasec.DStream.Plugin/IDStreamPlugin.cs`
  - Description: Update the interface to extend a common base interface

- [ ] **Task 4.2:** Update DStreamPluginHost for Compatibility
  - Project: Katasec.DStream.Plugin
  - File: `/src/Katasec.DStream.Plugin/DStreamPluginHost.cs`
  - Description: Update the host to handle both generic and non-generic plugins
  - Add logging to show which interface is being used

## Phase 5: Full Integration

- [ ] **Task 5.1:** Update Existing Plugins
  - Project: dstream-dotnet-test
  - File: `/CounterPlugin.cs`
  - Description: Update the existing plugin to use the base class
  - Add logging to show the plugin is using the typed configuration

- [ ] **Task 5.2:** Update Program Entry Point
  - Project: dstream-dotnet-test
  - File: `/Program.cs`
  - Description: Update the program to use the unified host

## Phase 6: Documentation and Examples

- [ ] **Task 6.1:** Update README
  - Project: Katasec.DStream.Plugin
  - File: `/readme.md`
  - Description: Update documentation to explain the generic configuration support

- [ ] **Task 6.2:** Create Example Plugins
  - Description: Create example plugins showing different configuration scenarios
    - Simple configuration
    - Nested configuration
    - Array/list configuration
    - Complex configuration with multiple types

## Notes

- This implementation plan allows for a gradual transition with minimal disruption
- Maintains backward compatibility while introducing the new strongly-typed approach
- Each phase can be tested independently before moving to the next

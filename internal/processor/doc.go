// Package processor provides functionality for processing database changes.
//
// The processor is responsible for:
// - Consuming changes from a message queue
// - Processing and transforming the changes as needed
// - Publishing the changes to configured destinations using the messaging package
//
// The main components are:
// - Processor: Core type that manages the processing lifecycle
// - Queue consumer: Reads changes from the message queue
// - Change processor: Processes and transforms changes
// - Publisher integration: Uses messaging package to publish changes
package processor

// Package router provides functionality for routing database change messages.
//
// The router is responsible for:
// - Consuming messages from the ingest queue
// - Reading the destination topic from the message metadata
// - Routing messages to their appropriate destination topics
//
// The main components are:
// - Router: Core type that manages the routing lifecycle
// - Queue consumer: Reads messages from the ingest queue
// - Message router: Routes messages to their destination topics
// - Publisher integration: Uses messaging package to publish to destination topics
package router

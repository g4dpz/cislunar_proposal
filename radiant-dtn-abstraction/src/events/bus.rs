//! Event bus implementation using tokio::broadcast.
//!
//! Provides a pub/sub system for distributing operational events
//! (link state changes, contact plan updates, engine state transitions)
//! to multiple concurrent subscribers.

use chrono::{DateTime, Utc};
use tokio::sync::broadcast;

use crate::lifecycle::EngineState;

/// Activity state of a link to a neighbor.
#[derive(Debug, Clone, Copy, PartialEq, Eq, serde::Serialize, serde::Deserialize)]
pub enum LinkActivity {
    /// Link is active and capable of carrying traffic.
    Active,
    /// Link is inactive (no connectivity).
    Inactive,
}

/// Action performed on the contact plan.
#[derive(Debug, Clone, Copy, PartialEq, Eq, serde::Serialize, serde::Deserialize)]
pub enum ContactAction {
    /// A contact was added to the plan.
    Added,
    /// A contact was removed from the plan.
    Removed,
}

/// Summary of a contact for event reporting.
#[derive(Debug, Clone, PartialEq, Eq, serde::Serialize, serde::Deserialize)]
pub struct ContactSummary {
    /// Source node number.
    pub source_node: u64,
    /// Destination node number.
    pub dest_node: u64,
    /// Contact start time (Unix timestamp seconds).
    pub start_time: i64,
    /// Contact end time (Unix timestamp seconds).
    pub end_time: i64,
}

/// Operational events published through the event bus.
///
/// Each variant includes a UTC timestamp and relevant details about
/// the state change that occurred.
#[derive(Debug, Clone, PartialEq, Eq, serde::Serialize)]
#[serde(tag = "type")]
pub enum DtnEvent {
    /// A link to a neighbor changed state (active/inactive).
    LinkStateChange {
        timestamp: DateTime<Utc>,
        neighbor_node: u64,
        link_id: String,
        new_state: LinkActivity,
    },
    /// The contact plan was modified (contact added or removed).
    ContactPlanChange {
        timestamp: DateTime<Utc>,
        action: ContactAction,
        contact: ContactSummary,
    },
    /// The DTN engine transitioned between lifecycle states.
    EngineStateChange {
        timestamp: DateTime<Utc>,
        old_state: EngineState,
        new_state: EngineState,
        reason: Option<String>,
    },
}

/// Broadcast-based event bus for distributing operational events
/// to multiple concurrent subscribers.
///
/// Uses `tokio::broadcast` internally, which supports multiple producers
/// and multiple consumers. Each subscriber receives all events published
/// after their subscription was created.
///
/// # Example
///
/// ```
/// use radiant_dtn_abstraction::events::bus::{EventBus, DtnEvent, LinkActivity};
/// use chrono::Utc;
///
/// let bus = EventBus::new(16);
/// let mut rx = bus.subscribe();
///
/// bus.publish(DtnEvent::LinkStateChange {
///     timestamp: Utc::now(),
///     neighbor_node: 20,
///     link_id: "ltp-to-orbiter".to_string(),
///     new_state: LinkActivity::Active,
/// });
/// ```
pub struct EventBus {
    sender: broadcast::Sender<DtnEvent>,
}

impl EventBus {
    /// Create a new EventBus with the given channel capacity.
    ///
    /// The capacity determines how many events can be buffered before
    /// slow subscribers start missing events (lagged).
    pub fn new(capacity: usize) -> Self {
        let (sender, _) = broadcast::channel(capacity);
        Self { sender }
    }

    /// Publish an event to all current subscribers.
    ///
    /// If there are no active subscribers, the event is silently dropped.
    pub fn publish(&self, event: DtnEvent) {
        // send() returns Err only when there are no receivers,
        // which is not an error condition for a pub/sub bus.
        let _ = self.sender.send(event);
    }

    /// Create a new subscriber that will receive all future events.
    ///
    /// The returned receiver will see events published after this call.
    /// If the subscriber falls behind by more than `capacity` events,
    /// it will receive a `RecvError::Lagged` on the next recv.
    pub fn subscribe(&self) -> broadcast::Receiver<DtnEvent> {
        self.sender.subscribe()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::Utc;

    #[tokio::test]
    async fn test_single_subscriber_receives_event() {
        let bus = EventBus::new(16);
        let mut rx = bus.subscribe();

        let event = DtnEvent::LinkStateChange {
            timestamp: Utc::now(),
            neighbor_node: 20,
            link_id: "ltp-to-orbiter".to_string(),
            new_state: LinkActivity::Active,
        };

        bus.publish(event);

        let received = rx.recv().await.unwrap();
        match received {
            DtnEvent::LinkStateChange {
                neighbor_node,
                link_id,
                new_state,
                ..
            } => {
                assert_eq!(neighbor_node, 20);
                assert_eq!(link_id, "ltp-to-orbiter");
                assert_eq!(new_state, LinkActivity::Active);
            }
            _ => panic!("Expected LinkStateChange event"),
        }
    }

    #[tokio::test]
    async fn test_multiple_subscribers_all_receive() {
        let bus = EventBus::new(16);
        let mut rx1 = bus.subscribe();
        let mut rx2 = bus.subscribe();
        let mut rx3 = bus.subscribe();

        let event = DtnEvent::EngineStateChange {
            timestamp: Utc::now(),
            old_state: EngineState::Stopped,
            new_state: EngineState::Starting,
            reason: Some("operator request".to_string()),
        };

        bus.publish(event);

        let r1 = rx1.recv().await.unwrap();
        let r2 = rx2.recv().await.unwrap();
        let r3 = rx3.recv().await.unwrap();

        // All three should have received the same event type
        for received in [&r1, &r2, &r3] {
            match received {
                DtnEvent::EngineStateChange {
                    old_state,
                    new_state,
                    reason,
                    ..
                } => {
                    assert_eq!(*old_state, EngineState::Stopped);
                    assert_eq!(*new_state, EngineState::Starting);
                    assert_eq!(reason.as_deref(), Some("operator request"));
                }
                _ => panic!("Expected EngineStateChange event"),
            }
        }
    }

    #[tokio::test]
    async fn test_contact_plan_change_event() {
        let bus = EventBus::new(16);
        let mut rx = bus.subscribe();

        let event = DtnEvent::ContactPlanChange {
            timestamp: Utc::now(),
            action: ContactAction::Added,
            contact: ContactSummary {
                source_node: 10,
                dest_node: 20,
                start_time: 1700000000,
                end_time: 1700003600,
            },
        };

        bus.publish(event);

        let received = rx.recv().await.unwrap();
        match received {
            DtnEvent::ContactPlanChange {
                action, contact, ..
            } => {
                assert_eq!(action, ContactAction::Added);
                assert_eq!(contact.source_node, 10);
                assert_eq!(contact.dest_node, 20);
                assert_eq!(contact.start_time, 1700000000);
                assert_eq!(contact.end_time, 1700003600);
            }
            _ => panic!("Expected ContactPlanChange event"),
        }
    }

    #[tokio::test]
    async fn test_publish_with_no_subscribers_does_not_panic() {
        let bus = EventBus::new(16);
        // No subscribers — should not panic
        bus.publish(DtnEvent::LinkStateChange {
            timestamp: Utc::now(),
            neighbor_node: 99,
            link_id: "test".to_string(),
            new_state: LinkActivity::Inactive,
        });
    }

    #[tokio::test]
    async fn test_event_timestamp_is_valid() {
        let bus = EventBus::new(16);
        let mut rx = bus.subscribe();

        let before = Utc::now();
        let event = DtnEvent::LinkStateChange {
            timestamp: Utc::now(),
            neighbor_node: 1,
            link_id: "test-link".to_string(),
            new_state: LinkActivity::Active,
        };
        bus.publish(event);
        let after = Utc::now();

        let received = rx.recv().await.unwrap();
        match received {
            DtnEvent::LinkStateChange { timestamp, .. } => {
                assert!(timestamp >= before);
                assert!(timestamp <= after);
            }
            _ => panic!("Expected LinkStateChange"),
        }
    }
}

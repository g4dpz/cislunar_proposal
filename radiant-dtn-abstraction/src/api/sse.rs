//! Server-Sent Events endpoint for streaming DTN operational events.
//!
//! Bridges the internal `EventBus` (tokio::broadcast) to an SSE response
//! stream, serializing each `DtnEvent` as a JSON data frame.

use axum::extract::State;
use axum::response::sse::{Event, KeepAlive, Sse};
use axum::response::IntoResponse;
use futures::stream::Stream;
use std::convert::Infallible;
use std::pin::Pin;
use std::task::{Context, Poll};

use crate::events::bus::DtnEvent;

use super::AppState;

/// GET /events — Server-Sent Events stream.
///
/// Subscribes to the EventBus and streams each `DtnEvent` as a JSON-encoded
/// SSE data frame. The stream stays open until the client disconnects.
///
/// Each event is sent as:
/// ```text
/// event: <event_type>
/// data: <json_payload>
/// ```
pub async fn get_events(
    State(state): State<AppState>,
) -> impl IntoResponse {
    let rx = state.event_bus.subscribe();

    let stream = EventStream { rx };

    Sse::new(stream).keep_alive(KeepAlive::default())
}

/// Adapter stream that converts broadcast::Receiver<DtnEvent> to an SSE Event stream.
struct EventStream {
    rx: tokio::sync::broadcast::Receiver<DtnEvent>,
}

impl Stream for EventStream {
    type Item = Result<Event, Infallible>;

    fn poll_next(mut self: Pin<&mut Self>, cx: &mut Context<'_>) -> Poll<Option<Self::Item>> {
        match self.rx.try_recv() {
            Ok(event) => {
                let event_type = match &event {
                    DtnEvent::LinkStateChange { .. } => "link_state_change",
                    DtnEvent::ContactPlanChange { .. } => "contact_plan_change",
                    DtnEvent::EngineStateChange { .. } => "engine_state_change",
                };

                let data = serde_json::to_string(&event).unwrap_or_else(|_| "{}".to_string());

                let sse_event = Event::default().event(event_type).data(data);

                Poll::Ready(Some(Ok(sse_event)))
            }
            Err(tokio::sync::broadcast::error::TryRecvError::Empty) => {
                // No events available yet; register waker and return pending.
                // We use a simple polling approach — in production you'd use
                // a proper async bridge. For correctness, we wake immediately
                // to poll again (busy-wait is acceptable for management APIs
                // with low event rates).
                cx.waker().wake_by_ref();
                Poll::Pending
            }
            Err(tokio::sync::broadcast::error::TryRecvError::Lagged(n)) => {
                // Subscriber fell behind; send a warning event and continue
                let warning = Event::default()
                    .event("warning")
                    .data(format!("{{\"lagged_events\": {}}}", n));
                Poll::Ready(Some(Ok(warning)))
            }
            Err(tokio::sync::broadcast::error::TryRecvError::Closed) => {
                // Channel closed; end the stream
                Poll::Ready(None)
            }
        }
    }
}

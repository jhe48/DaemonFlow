use serde::{Deserialize, Serialize};
use serde_json::Value;

// Request types (must match Go daemon constants)
pub const REQUEST_TYPE_PING: &str = "ping";
pub const REQUEST_TYPE_STATUS: &str = "status";
pub const REQUEST_TYPE_GET_STATE: &str = "get_state";
pub const REQUEST_TYPE_START_BREAK: &str = "start_break";
pub const REQUEST_TYPE_END_BREAK: &str = "end_break";
pub const REQUEST_TYPE_RESURRECT: &str = "resurrect";

#[derive(Debug, Serialize)]
pub struct Request {
    #[serde(rename = "type")]
    pub request_type: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub payload: Option<Value>,
}

#[derive(Debug, Deserialize)]
pub struct Response {
    pub success: bool,
    pub error: Option<String>,
    pub data: Option<Value>,
}

#[derive(Debug, Deserialize)]
pub struct StateResponse {
    pub clock_state: String,      // "working", "break", "overtime"
    pub earned_seconds: i32,      // Can be negative (overtime)
    pub session_earned: i32,      // Earned this session
    #[serde(default)]
    pub current_streak: i32,      // Days without death
    #[serde(default)]
    pub longest_streak: i32,      // Best streak ever
    #[serde(default)]
    pub total_deaths: i32,        // Total deaths
    #[serde(default = "default_level")]
    pub level: i32,               // Current pet level (1-5+)
    #[serde(default)]
    pub experience: i32,          // Current XP
    #[serde(default = "default_experience_to_next")]
    pub experience_to_next: i32,  // XP needed for next level
}

fn default_level() -> i32 {
    1
}

fn default_experience_to_next() -> i32 {
    100
}

#[derive(Debug, Deserialize)]
pub struct ClockEventResponse {
    pub previous_state: String,
    pub new_state: String,
}

#[derive(Debug, Deserialize)]
pub struct ResurrectResponse {
    pub success: bool,
    pub message: String,
}

// Task sync types
pub const REQUEST_TYPE_GET_TASKS: &str = "get_tasks";

#[derive(Debug, Deserialize)]
pub struct TaskData {
    pub id: i64,
    pub project_path: String,
    pub project_name: String,
    pub text: String,
    pub completed: bool,
    #[serde(default)]
    pub due_date: Option<String>,
    pub priority_score: i32,
    pub created_at: String,
}

#[derive(Debug, Deserialize)]
pub struct GetTasksResponse {
    pub tasks: Vec<TaskData>,
    pub total: i32,
}

use std::os::unix::net::UnixStream;
use std::io::{Read, Write};
use anyhow::{Context, Result, bail};
use super::protocol::*;

pub struct IpcClient {
    socket_path: String,
}

impl IpcClient {
    pub fn new() -> Self {
        // Default socket path: ~/.daemonflow/daemonflow.sock
        let home = dirs::home_dir().expect("Could not find home directory");
        let socket_path = home.join(".daemonflow").join("daemonflow.sock");
        Self { socket_path: socket_path.to_string_lossy().to_string() }
    }

    pub fn with_path(socket_path: String) -> Self {
        Self { socket_path }
    }

    fn send_request(&self, request: &Request) -> Result<Response> {
        // Connect to socket
        let mut stream = UnixStream::connect(&self.socket_path)
            .context("Failed to connect to daemon socket")?;

        // Serialize request to JSON
        let json = serde_json::to_vec(request)?;

        // Write 4-byte big-endian length prefix
        let len = json.len() as u32;
        stream.write_all(&len.to_be_bytes())?;

        // Write JSON payload
        stream.write_all(&json)?;

        // Read 4-byte length prefix for response
        let mut len_buf = [0u8; 4];
        stream.read_exact(&mut len_buf)?;
        let response_len = u32::from_be_bytes(len_buf) as usize;

        // Sanity check (max 1MB like Go daemon)
        if response_len > 1024 * 1024 {
            bail!("Response too large: {} bytes", response_len);
        }

        // Read response JSON
        let mut response_buf = vec![0u8; response_len];
        stream.read_exact(&mut response_buf)?;

        // Deserialize response
        let response: Response = serde_json::from_slice(&response_buf)?;
        Ok(response)
    }

    pub fn ping(&self) -> Result<bool> {
        let request = Request { request_type: REQUEST_TYPE_PING.to_string(), payload: None };
        let response = self.send_request(&request)?;
        Ok(response.success)
    }

    pub fn get_state(&self) -> Result<StateResponse> {
        let request = Request { request_type: REQUEST_TYPE_GET_STATE.to_string(), payload: None };
        let response = self.send_request(&request)?;
        if !response.success {
            bail!("get_state failed: {}", response.error.unwrap_or_default());
        }
        let data = response.data.context("No data in response")?;
        let state: StateResponse = serde_json::from_value(data)?;
        Ok(state)
    }

    pub fn start_break(&self) -> Result<ClockEventResponse> {
        let request = Request { request_type: REQUEST_TYPE_START_BREAK.to_string(), payload: None };
        let response = self.send_request(&request)?;
        if !response.success {
            bail!("start_break failed: {}", response.error.unwrap_or_default());
        }
        let data = response.data.context("No data in response")?;
        let event: ClockEventResponse = serde_json::from_value(data)?;
        Ok(event)
    }

    pub fn end_break(&self) -> Result<ClockEventResponse> {
        let request = Request { request_type: REQUEST_TYPE_END_BREAK.to_string(), payload: None };
        let response = self.send_request(&request)?;
        if !response.success {
            bail!("end_break failed: {}", response.error.unwrap_or_default());
        }
        let data = response.data.context("No data in response")?;
        let event: ClockEventResponse = serde_json::from_value(data)?;
        Ok(event)
    }

    pub fn is_daemon_running(&self) -> bool {
        self.ping().unwrap_or(false)
    }
}

impl Default for IpcClient {
    fn default() -> Self {
        Self::new()
    }
}

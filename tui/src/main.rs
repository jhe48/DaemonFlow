use anyhow::Result;
use daemonflow_tui::{app::App, terminal};

fn main() -> Result<()> {
    // Set up terminal
    let mut terminal = terminal::setup()?;

    // Create and run app
    let mut app = App::new();
    let result = app.run(&mut terminal);

    // Restore terminal (always, even on error)
    terminal::restore()?;

    result
}

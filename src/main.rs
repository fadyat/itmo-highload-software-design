use anyhow::Result;
use std::io::{Write, stdin, stdout};

fn main() -> Result<()> {
    loop {
        print!("> ");
        stdout().flush()?;

        let mut input = String::new();
        if stdin().read_line(&mut input).is_ok() {
            interpreter::run(&input)?;
        }
    }
}

pub mod executor;
pub mod lexer;
pub mod parser;
pub mod runtime;

use anyhow::Result;

pub fn run(input: &str) -> Result<()> {
    let tokens = lexer::lex(input);
    println!("Tokens: {:?}", tokens);

    let command = parser::parse(&tokens)?;
    println!("Command: {:?}", command);

    unimplemented!()
}

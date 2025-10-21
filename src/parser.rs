use crate::lexer::Token;
use anyhow::Result;

#[derive(Debug)]
pub struct CommandArgs {
    pub args: Vec<Token>,
    pub stdin: Option<Token>,
    pub stdout: Option<Token>,
}

#[derive(Debug)]
pub enum BuiltinType {
    Cat,
    Echo,
    Wc,
    Pwd,
    Exit,
}

#[derive(Debug)]
pub enum Command {
    Builtin { kind: BuiltinType, io: CommandArgs },
    SetVariable { name: String, value: Token },
    External { name: String, io: CommandArgs },
}

pub fn parse(tokens: &[Token]) -> Result<Command> {
    _ = tokens;
    unimplemented!()
}

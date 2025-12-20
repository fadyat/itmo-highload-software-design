use crate::lexer::Token;
use anyhow::{anyhow, Result};

#[derive(Debug)]
pub struct CommandArgs {
    pub args: Vec<Token>,
    pub stdin: Option<Token>,
    pub stdout: Option<Token>,
}

impl CommandArgs {
    pub fn new() -> Self {
        Self {
            args: Vec::new(),
            stdin: None,
            stdout: None,
        }
    }
}

#[derive(Debug)]
pub enum BuiltinType {
    Cat,
    Echo,
    Wc,
    Pwd,
    Exit,
}

fn to_builtin_type(text: &String) -> Option<BuiltinType> {
    match text.as_str() {
        "cat" => Some(BuiltinType::Cat),
        "echo" => Some(BuiltinType::Echo),
        "wc" => Some(BuiltinType::Wc),
        "pwd" => Some(BuiltinType::Pwd),
        "exit" => Some(BuiltinType::Exit),
        _ => None,
    }
}

fn push_token_to_cmd(cmd: &mut Command, token: Token) {
    match cmd {
        Command::Builtin { io, .. } | Command::External { io, .. } => {
            io.args.push(token);
        },
        _ => {}
    }
}

#[derive(Debug)]
pub enum Command {
    Builtin { kind: BuiltinType, io: CommandArgs },
    SetVariable { name: String, value: Token },
    External { name: String, io: CommandArgs },
}

pub fn parse(tokens: &[Token]) -> Result<Command> {
    let mut cur_cmd: Option<Command> = None;
    let mut stored_name: Option<String> = None;

    for token in tokens {
        match token {
            Token::Text(text) => {
                if cur_cmd.is_none() {
                    let cmd = if let Some(kind) = to_builtin_type(text) {
                        Command::Builtin {
                            kind,
                            io: CommandArgs::new(),
                        }
                    } else {
                        Command::External {
                            name: text.to_string(),
                            io: CommandArgs::new(),
                        }
                    };
                    stored_name = Some(text.to_string());
                    cur_cmd = Some(cmd);
                } else {
                    if let Some(cmd) = &mut cur_cmd {
                        push_token_to_cmd(cmd, token.clone());
                    }
                }
            }
            Token::QuotedText(text, _) => {
                let name = stored_name.take()
                    .ok_or_else(|| anyhow!("Variable name expected before quoted text"))?;

                cur_cmd = Some(Command::SetVariable {
                    name,
                    value: Token::Text(text.to_string()),
                });
            }
            Token::Stdin | Token::Stdout => {
                if let Some(cmd) = &mut cur_cmd {
                    push_token_to_cmd(cmd, token.clone());
                }
            }
            _ => {}
        }
    }

    cur_cmd.ok_or_else(|| anyhow!("No commands found in input"))
}
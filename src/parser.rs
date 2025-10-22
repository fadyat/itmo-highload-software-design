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
        Command::Builtin { io, .. } => {
            io.args.push(token);
        },
        Command::External { io, .. } => {
            io.args.push(token);
        },
        Command::SetVariable { value } => {
            *value = token;
        },
    }
}

#[derive(Debug)]
pub enum Command {
    Builtin { kind: BuiltinType, io: CommandArgs },
    SetVariable { value: Token },
    External { name: String, io: CommandArgs },
}

pub fn parse(tokens: &[Token]) -> Result<Command> {
    let mut cmds: Vec<Command> = Vec::new();
    let mut cur_cmd: Option<usize> = None;
    for token in tokens {
        match token {
            Token::Text(text) => {
                if cur_cmd.is_none() {
                    let command_type = match text.as_str() {
                        "cat" | "echo" | "wc" | "pwd" | "exit" =>
                            Command::Builtin {
                                kind: to_builtin_type(text).unwrap(),
                                io: CommandArgs {args: Vec::new(), stdin: None, stdout: None},
                            },
                        _ =>
                            Command::External {
                                name: text.to_string(),
                                io: CommandArgs {args: Vec::new(), stdin: None, stdout: None},
                            },
                    };
                    cmds.push(command_type);
                    cur_cmd = Some(cmds.len() - 1);
                } else {
                    if let Some(idx) = cur_cmd {
                        if let Some(cmd) = cmds.get_mut(idx) {
                            push_token_to_cmd(cmd, token.clone());
                        }
                    }
                }
            },
            Token::QuotedText(_, _) => {
                if let Some(idx) = cur_cmd {
                    if let Some(cmd) = cmds.get_mut(idx) {
                        push_token_to_cmd(cmd, token.clone());
                    }
                }
            },
            Token::Stdin | Token::Stdout | Token::Equal => {
                if let Some(idx) = cur_cmd {
                    if let Some(cmd) = cmds.get_mut(idx) {
                        push_token_to_cmd(cmd, token.clone());
                    }
                }
            },
            Token::Pipe => {
                cur_cmd = None;
            },
            _ => unimplemented!()
        }
    }

    unimplemented!()
}
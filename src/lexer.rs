/// Represents the state of quotation parsing.
/// Weak quotes allow for interpolation, while strong quotes do not.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum QuoteState {
    Weak,
    Strong,
}

impl From<char> for QuoteState {
    fn from(c: char) -> Self {
        match c {
            '"' => QuoteState::Weak,
            '\'' => QuoteState::Strong,
            _ => panic!("Invalid quote character"),
        }
    }
}

/// Represents the different types of tokens that can be identified in the input.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Token {
    Text(String),
    QuotedText(String, QuoteState),
    Stdin,
    Stdout,
    Equal,
}

/// Lexical analyzer that converts input string into a vector of tokens.
pub fn lex(input: &str) -> Vec<Token> {
    let mut tokens = Vec::new();
    let mut current_token = String::new();
    let mut chars = input.chars().peekable();

    let push_text_token = |tokens: &mut Vec<Token>, current_token: &mut String| {
        if !current_token.is_empty() {
            tokens.push(Token::Text(current_token.clone()));
            current_token.clear();
        }
    };

    while let Some(c) = chars.next() {
        match c {
            ' ' | '\n' | '\t' => push_text_token(&mut tokens, &mut current_token),
            '\'' | '"' => {
                // no escaping supported, for simplicity
                let content = chars.by_ref().take_while(|&ch| ch != c).collect::<String>();
                chars.next();
                tokens.push(Token::QuotedText(content, QuoteState::from(c)));
            }
            '>' => {
                push_text_token(&mut tokens, &mut current_token);
                tokens.push(Token::Stdout);
            }
            '<' => {
                push_text_token(&mut tokens, &mut current_token);
                tokens.push(Token::Stdin);
            }
            '=' => {
                push_text_token(&mut tokens, &mut current_token);
                tokens.push(Token::Equal);
            }
            _ => current_token.push(c),
        }
    }

    push_text_token(&mut tokens, &mut current_token);
    tokens
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_lex_simple_command() {
        assert_eq!(
            lex(r#"echo "Hello, World!" > output.txt"#),
            vec![
                Token::Text("echo".to_string()),
                Token::QuotedText("Hello, World!".to_string(), QuoteState::Weak),
                Token::Stdout,
                Token::Text("output.txt".to_string()),
            ]
        );
    }

    #[test]
    fn test_lex_with_assignment() {
        assert_eq!(
            lex(r#"VAR='value' ls -la < input.txt"#),
            vec![
                Token::Text("VAR".to_string()),
                Token::Equal,
                Token::QuotedText("value".to_string(), QuoteState::Strong),
                Token::Text("ls".to_string()),
                Token::Text("-la".to_string()),
                Token::Stdin,
                Token::Text("input.txt".to_string()),
            ]
        );
    }

    #[test]
    fn test_lex_with_multiple_spaces() {
        assert_eq!(
            lex(r#"   cmd    arg1   arg2   "#),
            vec![
                Token::Text("cmd".to_string()),
                Token::Text("arg1".to_string()),
                Token::Text("arg2".to_string()),
            ]
        );
    }

    #[test]
    fn test_lex_empty_input() {
        assert_eq!(lex(""), vec![]);
    }

    #[test]
    fn test_lex_only_special_chars() {
        assert_eq!(
            lex(r#"< > ="#),
            vec![Token::Stdin, Token::Stdout, Token::Equal,]
        );
    }
}

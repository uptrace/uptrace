interface FilterState {
  attr: string
  values: string[]
}

export function extractFilterState(s: string): FilterState | null {
  const lexer = new Lexer(s)

  let t = lexer.nextToken()
  if (t.text.toLowerCase() !== 'where') {
    return null
  }

  t = lexer.nextToken()
  if (t.id !== TokenID.Ident) {
    return null
  }

  const attr = t.text
  const values = []

  t = lexer.nextToken()
  switch (t.text.toLowerCase()) {
    case '=':
    case '==': {
      const value = nextValue(lexer)
      if (value !== '') {
        values.push(value)
      }
      break
    }
    case 'in': {
      t = lexer.nextToken()
      if (t.text !== '(') {
        return null
      }

      for (let i = 0; i < 10; i++) {
        if (i > 0) {
          t = lexer.nextToken()
          if (t.text !== ',') {
            break
          }
        }

        const value = nextValue(lexer)
        if (value !== '') {
          values.push(value)
        }
      }
      break
    }
    default:
      return null
  }

  if (values.length) {
    return { attr, values }
  }
  return null
}

function nextValue(lexer: Lexer): string {
  const t = lexer.nextToken()
  switch (t.id) {
    case TokenID.Ident:
    case TokenID.Value:
      return t.text
    default:
      return ''
  }
}

interface Token {
  id: TokenID
  text: string
}

export enum TokenID {
  Invalid = 0,
  Char,
  Ident,
  Value,
}

export const EOF = { id: TokenID.Invalid, text: '' }

export class Lexer {
  s = ''
  i = 0

  constructor(s: string) {
    this.s = s
  }

  nextToken(): Token {
    if (!this.isValid()) {
      return EOF
    }

    const ch = this.nextChar()

    switch (ch) {
      case "'":
      case '"':
        return this.quotedValue(ch)
    }

    if (isWhitespace(ch)) {
      return this.nextToken()
    }
    if (isAlpha(ch)) {
      return this.ident(this.i - 1)
    }
    if (isDigit(ch)) {
      return this.value(this.i - 1)
    }
    return { id: TokenID.Char, text: ch }
  }

  private quotedValue(quote: string): Token {
    const text = this.readUnquoted(quote)
    return { id: TokenID.Value, text }
  }

  private ident(start: number): Token {
    while (this.isValid()) {
      const ch = this.peekChar()
      if (!isIdent(ch)) {
        break
      }
      this.advance()
    }

    const text = this.s.slice(start, this.i)
    return { id: TokenID.Ident, text }
  }

  private value(start: number): Token {
    while (this.isValid()) {
      const ch = this.peekChar()
      if (isWhitespace(ch)) {
        break
      }
      this.advance()
    }

    const text = this.s.slice(start, this.i)
    return { id: TokenID.Value, text }
  }

  private readUnquoted(quote: string) {
    const buf = []

    while (this.isValid()) {
      const ch = this.nextChar()

      switch (ch) {
        case '\\': {
          const next = this.nextChar()
          switch (next) {
            case quote:
              buf.push(quote)
              break
            case '\\':
              buf.push('\\')
              break
            case 'n':
              buf.push('\n')
              break
            case 'r':
              buf.push('\r')
              break
            case 't':
              buf.push('\t')
              break
          }
          break
        }
        case quote:
          return buf.join('')
        default:
          buf.push(ch)
          break
      }
    }

    return buf.join('')
  }

  private isValid() {
    return this.i < this.s.length
  }

  nextChar() {
    const ch = this.s.charAt(this.i)
    this.i++
    return ch
  }

  private peekChar() {
    if (this.isValid()) {
      return this.s.charAt(this.i)
    }
    return ''
  }

  private advance() {
    this.i++
  }
}

function isAlpha(ch: string): boolean {
  return ch.toUpperCase() != ch.toLowerCase()
}

function isIdent(ch: string): boolean {
  switch (ch) {
    case '.':
    case '_':
      return true
    default:
      return isAlpha(ch)
  }
}

function isWhitespace(ch: string): boolean {
  return /\s/.test(ch)
}

function isDigit(ch: string): boolean {
  return /\d/.test(ch)
}

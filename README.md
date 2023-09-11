# ftail
Filtered Tail (ftail)

## Usage
```ftail <regex> <filename>```

## Description
This program simply follows a file (tail -f <filename>) and then executes a regex search across the newly inserted data (grep <expression>).

## Motivation
Throughout the hundreds of hours I have tailing files and looking for changes, I always needed a tool that only gave me the data I wanted. Welp, this is it now. The reason I chose golang as the language was primarily just to try out some golang. Sure this can be done with a small bash script, but it can also be done with a small golang program. So...

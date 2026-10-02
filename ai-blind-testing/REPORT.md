# Task 3 — AI blind test generation

## Function under test

`func ValidateEmail(s string) bool` from `validate/validate.go`

## Prompt you gave the AI

Write a table-driven Go test TestValidateEmail for this signature only:

func ValidateEmail(s string) bool

Do not look at my implementation. Give at least 8 cases: empty, missing @,
two @, spaces, unicode, trailing dot, consecutive dots, very long local part,
and a normal valid email. Return only the test table.

## Edge cases the AI found that you had missed

Two "@" in one string (`a@b@c.com`). I only counted "@" == 1 so it already
fails, but I did not put this case in my first table. No real bug.

## Edge cases you had that the AI missed

Phone numbers with dashes vs spaces. AI stayed on email only, which is fair
because the prompt was email-only. Also unicode in the local part — AI
mentioned it but expected true. I reject non-ascii.

## Cases where the AI's expected output was wrong

AI said `user@localhost` should be valid. My rule needs a dot in the domain,
so I keep it false. Also AI wanted trailing-dot emails to pass. I reject them.

## What you'd change about your own test-writing process after this

Ask the AI for cases first, then write the function. I still check every
expected value, because the model guessed a few of them wrong. Keep the
table in the test file, not only in chat.

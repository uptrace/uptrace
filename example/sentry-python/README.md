# Using sentry-python with Uptrace

To run this example:

```shell
python3 main.py
```

Then search for `RuntimeError` in your Uptrace project. The reported event
contains the full nested exception chain (`RuntimeError` caused by `KeyError`).

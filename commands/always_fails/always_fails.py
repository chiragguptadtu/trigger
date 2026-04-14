# ---trigger---
# name: Always Fails
# description: A command that always returns an error — used to test failure handling
# inputs:
#   - name: reason
#     label: Reason
#     type: open
#     required: true
# ---end---

import sys
import json

def run(inputs, config):
    reason = inputs.get("reason", "")
    return f"Intentional failure: {reason}"

if __name__ == "__main__":
    inputs = json.loads(sys.argv[1]) if len(sys.argv) > 1 else {}
    config = json.loads(sys.argv[2]) if len(sys.argv) > 2 else {}
    result = run(inputs, config)
    if result:
        print(result, file=sys.stderr)
        sys.exit(1)

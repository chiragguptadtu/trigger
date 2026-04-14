# ---trigger---
# name: Send Test Email
# description: Send a test email to verify email delivery is working
# inputs:
#   - name: recipient
#     label: Recipient Email
#     type: open
#     required: true
#   - name: template
#     label: Template
#     type: closed
#     options: [welcome, password_reset, invoice, notification]
#     multi: false
#     required: true
# ---end---

import sys
import json

def run(inputs, config):
    recipient = inputs.get("recipient")
    template = inputs.get("template")
    print(f"Sending {template} email to {recipient}")
    return ""

if __name__ == "__main__":
    inputs = json.loads(sys.argv[1]) if len(sys.argv) > 1 else {}
    config = json.loads(sys.argv[2]) if len(sys.argv) > 2 else {}
    result = run(inputs, config)
    if result:
        print(result, file=sys.stderr)
        sys.exit(1)

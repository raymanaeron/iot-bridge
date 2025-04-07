async function sendPrompt() {
    const input = document.getElementById("prompt");
    const chat = document.getElementById("chat");
    const text = input.value.trim();
    if (!text) return;
  
    // Show prompt
    const userMsg = document.createElement("div");
    userMsg.className = "user";
    userMsg.textContent = "You: " + text;
    chat.appendChild(userMsg);
  
    input.value = "";
  
    // Send to backend
    const res = await fetch("/llm", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ prompt: text })
    });
  
    const json = await res.json();
    const assistantMsg = document.createElement("div");
    assistantMsg.className = "assistant";
  
    assistantMsg.textContent = json.actions
      ? json.actions.map(a => `→ ${a.method} ${a.endpoint}\n   ${a.status}`).join("\n")
      : "⚠️ No actions";
  
    chat.appendChild(assistantMsg);
    chat.scrollTop = chat.scrollHeight;
  }
  
// Cache DOM elements
const chatLog = document.querySelector(".chat");
const chatWrapper = document.querySelector(".chat-wrapper");
const closeChatBtn = document.querySelector(".close-chat");
const chatInput = document.getElementById("chat-input");
let currentChatId = null; // Track the current chat user
let connection = null;    // Should be set to your WebSocket connection
let currentUserId = null; // Should be set to the logged-in user's ID

// Append messages with batch rendering
function appendMessages(messages, currentUserId) {
    const fragment = document.createDocumentFragment();
    const shouldScroll = chatLog.scrollTop > chatLog.scrollHeight - chatLog.clientHeight - 1;

    messages.forEach(({sender_id, content, date}) => {
        const container = document.createElement("div");
        const message = document.createElement("div");
        const time = document.createElement("div");

        const isSender = sender_id === currentUserId;
        container.className = isSender ? "sender-container" : "receiver-container";
        message.className = isSender ? "sender" : "receiver";
        message.textContent = content;
        time.className = "chat-time";
        time.textContent = date.slice(0, -3);

        container.append(message, time);
        fragment.appendChild(container);
    });

    chatLog.innerHTML = "";
    chatLog.appendChild(fragment);

    if (shouldScroll) {
        chatLog.scrollTop = chatLog.scrollHeight;
    }
}

// Fetch messages for a user
async function fetchMessages(receiverId) {
    try {
        const response = await fetch(`http://localhost:8000/message?receiver=${receiverId}`);
        return await response.json();
    } catch (error) {
        console.error("Failed to fetch messages:", error);
        return [];
    }
}

// Send a message and reload messages
async function sendMessage(connection, receiverId, messageInput, msgType, currentUserId) {
    if (!connection || !messageInput.value.trim()) return false;

    try {
        const msgData = {
            sender_id: currentUserId,
            receiver_id: receiverId,
            content: messageInput.value,
            msg_type: msgType,
            date: new Date().toISOString()
        };

        connection.send(JSON.stringify(msgData));
        messageInput.value = "";

        // Refresh messages after sending
        const messages = await fetchMessages(receiverId);
        appendMessages(messages, currentUserId);

        return true;
    } catch (error) {
        console.error("Failed to send message:", error);
        return false;
    }
}

// Setup chat UI and listeners
function setupChatUI() {
    // Send button click
    document.body.addEventListener("click", async (e) => {
        if (e.target.id === "send-btn" || e.target.closest("#send-btn")) {
            const msgInput = document.getElementById("chat-input");
            if (currentChatId && connection && currentUserId) {
                await sendMessage(connection, currentChatId, msgInput, 'msg', currentUserId);
            }
        }
    });

    // Enter key in chat input
    chatInput.addEventListener("keydown", async (e) => {
        if (e.key === "Enter") {
            e.preventDefault();
            if (currentChatId && connection && currentUserId) {
                await sendMessage(connection, chatInput, chatInput, 'msg', currentUserId);
            }
        }
    });

    // Close chat
    closeChatBtn.addEventListener("click", () => {
        chatWrapper.style.display = "none";
    });
}

// Open chat and load messages
async function openChat(receiverId, wsConnection, userId) {
    try {
        currentChatId = receiverId;
        connection = wsConnection;
        currentUserId = userId;

        // Update UI
        const user = allUsers.find(u => u.id === receiverId);
        if (user) {
            document.querySelector(".chat-user-username").textContent = user.username;
        }
        chatWrapper.style.display = "flex";

        // Mark as read (if you have unread logic)
        if (typeof unread !== "undefined") {
            const unreadIndex = unread.findIndex(u => u[0] === receiverId);
            if (unreadIndex !== -1) {
                unread[unreadIndex][1] = 0;
            }
        }

        // Load messages
        const messages = await fetchMessages(receiverId);
        appendMessages(messages, currentUserId);
    } catch (error) {
        console.error("Failed to open chat:", error);
    }
}

// Poll for new messages every 2 seconds
setInterval(async () => {
    if (currentChatId && currentUserId) {
        const messages = await fetchMessages(currentChatId);
        appendMessages(messages, currentUserId);
    }
}, 2000);

// Initialize listeners
setupChatUI();
import { getData, DOM, currId, allUsers, unread, conn, API_BASE_URL } from './index.js';
import { OpenChat } from './message.js';

async function initializeChat(apiUrl, allUsers, online, conn, DOM) {
  try {
    allUsers.length = 0;
    const users = await getData(`${apiUrl}/user`);
    allUsers.push(...users);
    await updateUsers(apiUrl, allUsers, online, conn, DOM);
  } catch (error) {
    console.error('Failed to initialize chat:', error);
  }
}

function createUsers(users, online, conn, DOM, currId) {
  DOM.onlineUsers.innerHTML = '';
  DOM.offlineUsers.innerHTML = '';

  if (!users || users.length === 0) {
    return;
  }

  users.forEach(({ id, username }) => {
    if (id === currId) return;

    const userElement = document.createElement('div');
    userElement.className = 'user';
    userElement.setAttribute('id', `id${id}`);
    userElement.innerHTML = `
      <img src="./frontend/assets/profile4.svg" alt="Profile">
      <p>${username}</p>
      <div class="msg-notification">1</div>
    `;

    const targetContainer = online.includes(id) ? DOM.onlineUsers : DOM.offlineUsers;
    targetContainer.appendChild(userElement);

    const unreadMsgs = unread.find(u => u[0] === id);
    if (unreadMsgs && unreadMsgs[1] > 0) {
      const msgNotif = userElement.querySelector('.msg-notification');
      msgNotif.style.opacity = '1';
      msgNotif.innerText = unreadMsgs[1];
      userElement.style.fontWeight = '900';
    }

    userElement.addEventListener('click', () => handleUserClick(id, apiUrl, conn, DOM, currId));
  });
}

async function updateUsers(apiUrl, allUsers, online, conn, DOM) {
  try {
    const { user_ids } = await getData(`${apiUrl}/chat?user_id=${currId}`);
    const newUsers = user_ids ? user_ids.map(id => allUsers.find(u => u.id === id)).filter(Boolean) : [];
    const otherUsers = allUsers.filter(u => !newUsers.includes(u));
    allUsers.length = 0;
    allUsers.push(...newUsers, ...otherUsers);
    createUsers(allUsers, online, conn, DOM, currId);
  } catch (error) {
    console.error('Failed to update users:', error);
  }
}

function startWebSocket(wsUrl, { handleMessage }) {
  if (!window.WebSocket) {
    const item = document.createElement('div');
    item.innerHTML = '<b>Your browser does not support WebSockets.</b>';
    DOM.chatLog.appendChild(item);
    return null;
  }

  const ws = new WebSocket(wsUrl);
  ws.onclose = () => {
    console.log('WebSocket connection closed.');
  };
  ws.onmessage = (evt) => {
    const data = JSON.parse(evt.data);
    handleMessage(data);
  };
  return ws;
}

function closeWebSocket(conn) {
  if (conn?.readyState === WebSocket.OPEN) {
    conn.close();
  }
}

function appendMessage(data, currId, chatLog, chatWrapper, unread) {
  const container = document.createElement('div');
  container.className = data.sender_id === currId ? 'sender-container' : 'receiver-container';
  const message = document.createElement('div');
  message.className = data.sender_id === currId ? 'sender' : 'receiver';
  message.innerText = data.content;
  const date = document.createElement('div');
  date.className = 'chat-time';
  date.innerText = data.date.slice(0, -3);

  const doScroll = chatLog.scrollTop > chatLog.scrollHeight - chatLog.clientHeight - 1;
  chatLog.appendChild(container);
  container.appendChild(message);
  message.appendChild(date);

  if (doScroll) {
    chatLog.scrollTop = chatLog.scrollHeight - chatLog.clientHeight;
  }

  if (data.sender_id !== currId) {
    const unreadMsgs = unread.find(u => u[0] === data.sender_id);
    if (chatWrapper.style.display === 'none') {
      if (!unreadMsgs) {
        unread.push([data.sender_id, 1]);
      } else {
        unreadMsgs[1] += 1;
      }
    }
  }
}

async function handleUserClick(receiverId, apiUrl, conn, DOM, currId) {
  try {
    const messages = await getData(`${apiUrl}/message?receiver=${receiverId}`);
    const msgNotif = document.getElementById(`id${receiverId}`)?.querySelector('.msg-notification');
    if (msgNotif) {
      msgNotif.style.opacity = '0';
    }
    const userElement = document.getElementById(`id${receiverId}`);
    if (userElement) {
      userElement.style.fontWeight = '400';
    }
    const unreadEntry = unread.find(u => u[0] === receiverId);
    if (unreadEntry) {
      unreadEntry[1] = 0;
    }
    OpenChat(receiverId, conn, messages, currId, DOM, allUsers, apiUrl);
  } catch (error) {
    console.error('Failed to fetch messages:', error);
  }
}

function sendMessage(conn, receiverId, message, type) {
  if (!conn?.readyState === WebSocket.OPEN || !message.value) {
    return false;
  }

  const msgData = {
    id: 0,
    sender_id: 0,
    receiver_id: receiverId,
    content: message.value,
    date: '',
    msg_type: type,
  };

  try {
    conn.send(JSON.stringify(msgData));
    message.value = '';
    updateUsers(API_BASE_URL, allUsers, [], conn, DOM);
    return true;
  } catch (error) {
    console.error('Failed to send message:', error);
    return false;
  }
}

function closeChat(DOM) {
  DOM.chatWrapper.style.display = 'none';
}

export { initializeChat, createUsers, updateUsers, startWebSocket, closeWebSocket, appendMessage, handleUserClick, sendMessage, closeChat };
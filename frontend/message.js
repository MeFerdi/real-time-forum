import { postData, getData, DOM, currComments } from './index.js';
import { createComments } from './post.js';
import { sendMessage, updateUsers } from './chat.js';

async function sendComment(postId, userId, apiUrl, DOM, currComments) {
  const content = DOM.commentInput.value.trim();
  if (!content) return;

  const commentData = {
    id: 0,
    post_id: postId,
    user_id: userId,
    content,
    date: '',
  };

  try {
    await postData(`${apiUrl}/comment`, commentData);
    DOM.commentInput.value = '';
    await getComments(postId, apiUrl);
    DOM.postComments.innerHTML = currComments?.length || 0 + ' Comments';
    createComments(currComments, DOM);
  } catch (error) {
    console.error('Failed to send comment:', error);
  }
}

async function getComments(postId, apiUrl) {
  try {
    currComments.length = 0;
    const comments = await getData(`${apiUrl}/comment?param=post_id&data=${postId}`);
    currComments.push(...comments);
  } catch (error) {
    console.error('Failed to fetch comments:', error);
  }
}

function OpenChat(receiverId, conn, messages, currId, DOM, allUsers, apiUrl) {
  DOM.chatWrapper.style.display = 'flex';
  DOM.chatLog.innerHTML = '';

  const username = allUsers.find(u => u.id === receiverId)?.username || 'Unknown';
  const chatUsernameElement = document.querySelector('.chat-user-username');
  if (chatUsernameElement) {
    chatUsernameElement.innerText = username;
  }

  if (messages && messages.length > 0) {
    messages.forEach(({ sender_id, content, date }) => {
      const container = document.createElement('div');
      container.className = sender_id === currId ? 'sender-container' : 'receiver-container';
      const message = document.createElement('div');
      message.className = sender_id === currId ? 'sender' : 'receiver';
      message.innerText = content;
      const messageDate = document.createElement('div');
      messageDate.className = 'chat-time';
      messageDate.innerText = date.slice(0, -3);

      DOM.chatLog.appendChild(container);
      container.appendChild(message);
      message.appendChild(messageDate);
    });
  }

  // Replace send-wrapper to reset event listeners
  const oldSendWrapper = document.querySelector('.send-wrapper');
  if (oldSendWrapper) {
    const newSendWrapper = oldSendWrapper.cloneNode(true);
    oldSendWrapper.parentNode.replaceChild(newSendWrapper, oldSendWrapper);
  }

  // Add event listeners for sending messages
  const sendBtn = document.querySelector('#send-btn');
  if (sendBtn) {
    sendBtn.addEventListener('click', async () => {
      if (sendMessage(conn, receiverId, DOM.chatInput, 'msg')) {
        try {
          const updatedMessages = await getData(`${apiUrl}/message?receiver=${receiverId}`);
          OpenChat(receiverId, conn, updatedMessages, currId, DOM, allUsers, apiUrl);
        } catch (error) {
          console.error('Failed to refresh messages:', error);
        }
      }
    });
  }

  DOM.chatInput.addEventListener('keydown', async (evt) => {
    if (evt.keyCode === 13) {
      if (sendMessage(conn, receiverId, DOM.chatInput, 'msg')) {
        try {
          const updatedMessages = await getData(`${apiUrl}/message?receiver=${receiverId}`);
          OpenChat(receiverId, conn, updatedMessages, currId, DOM, allUsers, apiUrl);
        } catch (error) {
          console.error('Failed to refresh messages:', error);
        }
      }
    }
  });

  DOM.chatInput.focus();
}

export { sendComment, OpenChat };
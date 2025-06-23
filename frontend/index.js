import { initializeAuth, handleSignIn, handleSignUp, handleLogout } from './auth.js';
import { initializePosts, createPosts, handleNewPost, handleCategoryChange, handleHome, handleNewPostNotification } from './post.js';
import { initializeChat, updateUsers, startWebSocket, closeWebSocket, appendMessage, closeChat } from './chat.js';
import { sendComment } from './message.js';

const API_BASE_URL = 'http://localhost:8000';
const WS_URL = 'ws://' + document.location.host + '/ws';

let conn = null;
let currId = 0;
let currUsername = '';
let currPost = 0;
let unread = [];
let allPosts = [];
let allUsers = [];
let currComments = [];
let online = [];


const DOM = {
  postsContainer: document.querySelector('.posts-container'),
  createPostContainer: document.querySelector('.create-post-container'),
  postContainer: document.querySelector('.post-container'),
  contentWrapper: document.querySelector('.content-wrapper'),
  registerContainer: document.querySelector('.register-container'),
  signinContainer: document.querySelector('.signin'),
  signupNav: document.querySelector('.signup-nav'),
  logoutNav: document.querySelector('.logout-nav'),
  onlineUsers: document.querySelector('.online-users'),
  offlineUsers: document.querySelector('.offline-users'),
  commentsContainer: document.querySelector('.comments-container'),
  topPanel: document.querySelector('.top-panel'),
  newPostNotif: document.querySelector('.new-post-notif-wrapper'),
  msgNotif: document.querySelector('.msg-notification'),
  chatWrapper: document.querySelector('.chat-wrapper'),
  chatInput: document.getElementById('chat-input'),
  chatLog: document.querySelector('.chat'),
  categories: document.getElementById('categories'),
  likeBtn: document.getElementById('like-btn'),
  dislikeBtn: document.getElementById('dislike-btn'),
  signinBtn: document.querySelector('.signin-btn'),
  signupBtn: document.querySelector('.signup-btn'),
  signupLink: document.querySelector('#signup-link'),
  signinLink: document.querySelector('#signin-link'),
  registerBtn: document.querySelector('.register-btn'),
  newPostBtn: document.querySelector('.new-post-btn'),
  createPostBtn: document.querySelector('.create-post-btn'),
  sendCommentBtn: document.querySelector('.send-comment-btn'),
  commentInput: document.querySelector('#comment-input'),
  logo: document.querySelector('.logo'),
  back: document.querySelector('.back'),
  backBtn: document.querySelector('#back-btn'),
  profile: document.querySelector('.profile'),
  closeChat: document.querySelector('.close-chat'),
};

async function initializeApp() {
  DOM.chatWrapper.style.display = 'none';

  try {
    await initializePosts(API_BASE_URL);
    await initializeAuth(API_BASE_URL, DOM, {
      setCurrId: (id) => (currId = id),
      setCurrUsername: (username) => (currUsername = username),
      startWebSocket: () => {
        conn = startWebSocket(WS_URL, {
          handleMessage: (data) => {
            if (data.msg_type === 'msg') {
              appendMessage(data, currId, DOM.chatLog, DOM.chatWrapper, unread);
              updateUsers(API_BASE_URL, allUsers, online, conn, DOM);
            } else if (data.msg_type === 'online') {
              online = data.user_ids;
              updateUsers(API_BASE_URL, allUsers, online, conn, DOM);
            } else if (data.msg_type === 'post') {
              DOM.newPostNotif.style.display = 'flex';
            }
          },
        });
      },
      createPosts: () => createPosts(allPosts, DOM, currPost, currComments, allUsers, API_BASE_URL),
      updateUsers: () => updateUsers(API_BASE_URL, allUsers, online, conn, DOM),
    });
    await initializeChat(API_BASE_URL, allUsers, online, conn, DOM);
  } catch (error) {
    console.error('Failed to initialize app:', error);
  }

  // Event Listeners
  DOM.categories.addEventListener('change', () => handleCategoryChange(allPosts, DOM));
  DOM.likeBtn.addEventListener('click', () => handleLikeDislike(currPost, 'likes', API_BASE_URL, DOM));
  DOM.dislikeBtn.addEventListener('click', () => handleLikeDislike(currPost, 'dislikes', API_BASE_URL, DOM));
  DOM.signinBtn.addEventListener('click', () => handleSignIn(API_BASE_URL, DOM, {
    setCurrId: (id) => (currId = id),
    setCurrUsername: (username) => (currUsername = username),
    startWebSocket: () => {
      conn = startWebSocket(WS_URL, {
        handleMessage: (data) => {
          if (data.msg_type === 'msg') {
            appendMessage(data, currId, DOM.chatLog, DOM.chatWrapper, unread);
            updateUsers(API_BASE_URL, allUsers, online, conn, DOM);
          } else if (data.msg_type === 'online') {
            online = data.user_ids;
            updateUsers(API_BASE_URL, allUsers, online, conn, DOM);
          } else if (data.msg_type === 'post') {
            DOM.newPostNotif.style.display = 'flex';
          }
        },
      });
    },
    createPosts: () => createPosts(allPosts, DOM, currPost, currComments, allUsers, API_BASE_URL),
    updateUsers: () => updateUsers(API_BASE_URL, allUsers, online, conn, DOM),
  }));
  DOM.signupLink.addEventListener('click', () => toggleAuthContainers(DOM, true));
  DOM.signinLink.addEventListener('click', () => toggleAuthContainers(DOM, false));
  DOM.signupBtn.addEventListener('click', () => toggleAuthContainers(DOM, DOM.signupBtn.innerText === 'SIGN UP'));
  DOM.registerBtn.addEventListener('click', (e) => handleSignUp(e, API_BASE_URL, DOM));
  DOM.newPostBtn.addEventListener('click', () => {
    DOM.postsContainer.style.display = 'none';
    DOM.postContainer.style.display = 'none';
    DOM.createPostContainer.style.display = 'flex';
    DOM.topPanel.style.display = 'none';
    document.querySelector('#create-post-title').value = '';
    document.querySelector('#create-post-body').value = '';
  });
  DOM.createPostBtn.addEventListener('click', () =>
    handleNewPost(API_BASE_URL, DOM, conn, allPosts, () => createPosts(allPosts, DOM, currPost, currComments, allUsers, API_BASE_URL))
  );
  DOM.sendCommentBtn.addEventListener('click', () => sendComment(currPost, currId, API_BASE_URL, DOM, currComments));
  DOM.commentInput.addEventListener('keydown', (event) => {
    if (event.keyCode === 13) {
      sendComment(currPost, currId, API_BASE_URL, DOM, currComments);
    }
  });
  DOM.logo.addEventListener('click', () => handleHome(API_BASE_URL, DOM, allPosts));
  DOM.back.addEventListener('click', () => handleHome(API_BASE_URL, DOM, allPosts));
  DOM.backBtn.addEventListener('click', () => handleHome(API_BASE_URL, DOM, allPosts));
  DOM.newPostNotif.addEventListener('click', () => handleNewPostNotification(API_BASE_URL, DOM, allPosts));
  DOM.closeChat.addEventListener('click', () => closeChat(DOM));
  DOM.logoutBtn.addEventListener('click', () => handleLogout(API_BASE_URL, DOM, () => closeWebSocket(conn)));
}

function toggleAuthContainers(DOM, showRegister) {
  DOM.signinContainer.style.display = showRegister ? 'none' : 'flex';
  DOM.registerContainer.style.display = showRegister ? 'block' : 'none';
  DOM.signupBtn.innerText = showRegister ? 'SIGN IN' : 'SIGN UP';
}

function handleLikeDislike(postId, column, apiUrl, DOM) {
  postData(`${apiUrl}/like?post_id=${postId}&col=${column}`)
    .then((data) => {
      const [likes, dislikes] = data.msg.split('|').map(Number);
      DOM.postLikes.innerHTML = likes;
      DOM.postDislikes.innerHTML = dislikes;
    })
    .catch((error) => console.error(`Failed to update ${column}:`, error));
}
async function postData(url = '', data = {}) {
    const response = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
        credentials: 'include'
    });
    return response.json();
}
async function getData(url = '') {
    const response = await fetch(url, {
        method: 'GET',
        credentials: 'include'
    });
    return response.json();
}

window.addEventListener('DOMContentLoaded', initializeApp);

export { postData, getData, API_BASE_URL, DOM, currId, currUsername, currPost, unread, allPosts, allUsers, currComments, conn };
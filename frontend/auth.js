import { postData, getData, DOM, currId, currUsername, allPosts, allUsers, conn } from './index.js';
import { createPosts, initializePosts } from './post.js';
import { updateUsers, startWebSocket, closeWebSocket } from './chat.js';

async function initializeAuth(apiUrl, DOM, { setCurrId, setCurrUsername, startWebSocket, createPosts, updateUsers }) {
  try {
    const { msg } = await postData(`${apiUrl}/session`);
    const [id, username] = msg.split('|');
    setCurrId(parseInt(id));
    setCurrUsername(username);
    DOM.profile.innerText = username;
    DOM.signinContainer.style.display = 'none';
    DOM.signupNav.style.display = 'none';
    DOM.contentWrapper.style.display = 'flex';
    DOM.logoutNav.style.display = 'flex';
    startWebSocket();
    await initializePosts(apiUrl);
    createPosts();
    await updateUsers();
  } catch (error) {
    DOM.signinContainer.style.display = 'flex';
    DOM.signupNav.style.display = 'flex';
    DOM.contentWrapper.style.display = 'none';
    DOM.logoutNav.style.display = 'none';
    console.error('Session check failed:', error);
  }
}

async function handleSignIn(apiUrl, DOM, { setCurrId, setCurrUsername, startWebSocket, createPosts, updateUsers }) {
  const emailUsername = document.querySelector('#email-username').value;
  const password = document.querySelector('#signin-password').value;

  if (!emailUsername || !password) {
    alert('Please enter both email/username and password.');
    return;
  }

  try {
    const { msg } = await postData(`${apiUrl}/login`, { emailUsername, password });
    const [id, username] = msg.split('|');
    setCurrId(parseInt(id));
    setCurrUsername(username);
    DOM.profile.innerText = username;
    DOM.signinContainer.style.display = 'none';
    DOM.signupNav.style.display = 'none';
    DOM.contentWrapper.style.display = 'flex';
    DOM.logoutNav.style.display = 'flex';
    document.querySelector('#email-username').value = '';
    document.querySelector('#signin-password').value = '';
    startWebSocket();
    await initializePosts(apiUrl);
    createPosts();
    await updateUsers();
  } catch (error) {
    console.error('Sign-in failed:', error);
    alert('Invalid credentials.');
  }
}

async function handleSignUp(event, apiUrl, DOM) {
  event.preventDefault();

  const fname = document.querySelector('#fname').value;
  const lname = document.querySelector('#lname').value;
  const email = document.querySelector('#email').value;
  const username = document.querySelector('#register-username').value;
  const age = document.querySelector('#age').value;
  const gender = document.querySelector('#gender').value;
  const password = document.querySelector('#register-password').value;

  const errors = [];
  if (!fname) errors.push('Enter a first name.');
  if (!lname) errors.push('Enter a last name.');
  if (!email) errors.push('Enter an email.');
  if (!username) errors.push('Enter a username.');
  if (!age) errors.push('Enter a date of birth.');
  if (!gender) errors.push('Select a gender.');
  if (!password) errors.push('Enter a password.');

  if (errors.length > 0) {
    alert(errors.join(' '));
    return;
  }

  const data = {
    id: 0,
    username,
    firstname: fname,
    surname: lname,
    gender,
    email,
    dob: age,
    password,
  };

  try {
    const { msg } = await postData(`${apiUrl}/register`, data);
    alert(msg);
    DOM.registerContainer.style.display = 'none';
    DOM.signinContainer.style.display = 'flex';
  } catch (error) {
    console.error('Registration failed:', error);
    alert('Registration failed.');
  }
}

async function handleLogout(apiUrl, DOM, closeWebSocket) {
  try {
    const { msg } = await postData(`${apiUrl}/logout`);
    console.log(msg);
    DOM.signinContainer.style.display = 'flex';
    DOM.registerContainer.style.display = 'none';
    DOM.contentWrapper.style.display = 'none';
    DOM.signupNav.style.display = 'flex';
    DOM.logoutNav.style.display = 'none';
    closeWebSocket();
  } catch (error) {
    console.error('Logout failed:', error);
    alert('Logout failed.');
  }
}

export { initializeAuth, handleSignIn, handleSignUp, handleLogout };
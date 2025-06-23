import { getData, postData, DOM, allPosts, currPost, currComments, allUsers, API_BASE_URL } from './index.js';

async function initializePosts(apiUrl) {
  try {
    allPosts.length = 0;
    const posts = await getData(`${apiUrl}/post`);
    allPosts.push(...posts);
  } catch (error) {
    console.error('Failed to fetch posts:', error);
  }
}

function createPosts(posts, DOM, currPost, currComments, allUsers, apiUrl) {
  DOM.postsContainer.innerHTML = '';

  if (!posts || posts.length === 0) {
    return;
  }

  posts.forEach(async (post) => {
    await getComments(post.id, apiUrl);

    const postElement = document.createElement('div');
    postElement.className = 'post';
    postElement.setAttribute('id', post.id);

    postElement.innerHTML = `
      <div class="title">${post.title}</div>
      <div class="author">
        <img src="./frontend/assets/profile7.svg" alt="Profile">
        <div class="post-username">${allUsers.find(u => u.id === post.user_id)?.username || 'Unknown'}</div>
        <div class="date">${post.date.slice(0, -3)}</div>
      </div>
      <div class="post-body">${post.content}</div>
      <div class="comments-wrapper">
        <div class="likes-dislikes-wrapper">
          <div class="likes-wrapper">
            <img src="./frontend/assets/like3.svg" alt="Like">
            <div class="likes">${post.likes}</div>
          </div>
          <div class="likes-wrapper dislike">
            <img src="./frontend/assets/dislike4.svg" alt="Dislike">
            <div class="dislike">${post.dislikes}</div>
          </div>
        </div>
        <div class="comments">
          <img src="./frontend/assets/comment.svg" alt="Comment">
          <div class="comment">${currComments?.length || 0} Comments</div>
        </div>
      </div>
    `;

    postElement.addEventListener('click', async () => {
      currPost = parseInt(postElement.getAttribute('id'));
      await getComments(currPost, apiUrl);
      createPost(posts.find(p => p.id === currPost), DOM, allUsers);
      createComments(currComments, DOM, allUsers);
      DOM.postComments.innerHTML = currComments?.length || 0 + ' Comments';
      DOM.postsContainer.style.display = 'none';
      DOM.postContainer.style.display = 'flex';
      DOM.topPanel.style.display = 'none';
    });

    DOM.postsContainer.appendChild(postElement);
  });
}

function createPost(post, DOM, allUsers) {
  if (!post) return;

  DOM.postTitle.innerHTML = post.title;
  DOM.postUsername.innerHTML = allUsers.find(u => u.id === post.user_id)?.username || 'Unknown';
  DOM.postDate.innerHTML = post.date.slice(0, -3);
  DOM.postCategory.innerHTML = post.category;
  DOM.postContent.innerHTML = post.content;
  DOM.postLikes.innerHTML = post.likes;
  DOM.postDislikes.innerHTML = post.dislikes;
}

function createComments(comments, DOM, allUsers) {
  DOM.commentsContainer.innerHTML = '';

  if (!comments || comments.length === 0) {
    return;
  }

  comments.forEach(({ id, user_id, content, date }) => {
    const commentWrapper = document.createElement('div');
    commentWrapper.className = 'comment-wrapper';
    commentWrapper.innerHTML = `
      <img src="./frontend/assets/profile7.svg" alt="Profile">
      <div class="comment">
        <div class="comment-user-wrapper">
          <div class="comment-username">${allUsers.find(u => u.id === user_id)?.username || 'Unknown'}</div>
          <div class="comment-date">${date.slice(0, -3)}</div>
        </div>
        <div>${content}</div>
      </div>
    `;
    DOM.commentsContainer.appendChild(commentWrapper);
  });
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

function handleCategoryChange(posts, DOM) {
  const category = DOM.categories.value;
  if (category === 'all') {
    createPosts(posts, DOM, currPost, currComments, allUsers, API_BASE_URL);
    return;
  }
  const filteredPosts = posts.filter(post => post.category === category);
  createPosts(filteredPosts, DOM, currPost, currComments, allUsers, API_BASE_URL);
}

async function handleNewPost(apiUrl, DOM, conn, posts, createPostsCallback) {
  const title = document.querySelector('#create-post-title').value;
  const body = document.querySelector('#create-post-body').value;
  const category = document.querySelector('#create-post-categories').value;

  if (!title || !body || !category) {
    alert('Please fill in all fields.');
    return;
  }

  const data = {
    id: 0,
    user_id: 0,
    category,
    title,
    content: body,
    date: '',
    likes: 0,
    dislikes: 0,
  };

  try {
    await postData(`${apiUrl}/post`, data);
    await initializePosts(apiUrl);
    createPostsCallback();
    sendMessage(conn, 0, { value: 'New Post' }, 'post');
    DOM.createPostContainer.style.display = 'none';
    DOM.postsContainer.style.display = 'flex';
    DOM.topPanel.style.display = 'flex';
  } catch (error) {
    console.error('Failed to create post:', error);
    alert('Failed to create post.');
  }
}

async function handleHome(apiUrl, DOM, posts) {
  DOM.categories.selectedIndex = 0;
  await initializePosts(apiUrl);
  createPosts(posts, DOM, currPost, currComments, allUsers, apiUrl);
  DOM.createPostContainer.style.display = 'none';
  DOM.postContainer.style.display = 'none';
  DOM.postsContainer.style.display = 'flex';
  DOM.topPanel.style.display = 'flex';
  DOM.newPostNotif.style.display = 'none';
}

async function handleNewPostNotification(apiUrl, DOM, posts) {
  await initializePosts(apiUrl);
  createPosts(posts, DOM, currPost, currComments, allUsers, apiUrl);
  DOM.newPostNotif.style.display = 'none';
  window.scrollTo(0, 0);
}

export { initializePosts, createPosts, createPost, createComments, getComments, handleCategoryChange, handleNewPost, handleHome, handleNewPostNotification };
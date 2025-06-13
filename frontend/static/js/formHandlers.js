import authService from './authService.js';
import postService from './postService.js';
import { uiComponents } from './uiComponents.js';

const formHandlers = {
    ws: null,

    async init() {
        this.initWebSocket();
        const route = window.location.hash.replace('#', '') || 'home';
        await this.handleRoute(route);
        window.addEventListener('hashchange', async () => {
            const newRoute = window.location.hash.replace('#', '') || 'home';
            await this.handleRoute(newRoute);
        });
    },

    async handleRoute(route) {
        if (route === 'signup') {
            this.initSignupForm();
        } else if (route === 'login') {
            this.initLoginForm();
        } else if (route === 'home') {
            await this.initHome();
        } else if (route === 'profile') {
            await this.initProfile();
        }
    },

    initWebSocket() {
        const token = authService.getToken();
        if (!token) return;
        const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        this.ws = new WebSocket(`${wsProtocol}//${window.location.host}/ws?token=${token}`);
        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                if (message.Type === 'post_created') {
                    this.appendPost(message.Data);
                }
            } catch (error) {
                console.error('WebSocket message parsing error:', error);
            }
        };
        this.ws.onclose = () => {
            console.log('WebSocket closed. Reconnecting...');
            setTimeout(() => this.initWebSocket(), 5000);
        };
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    },

    appendPost(post) {
        const feed = document.getElementById('feed');
        if (!feed) return;
        const activeCategoryId = document.querySelector('.category-item.active')?.dataset.categoryId || '0';
        if (activeCategoryId === '0' || post.category_ids?.includes(parseInt(activeCategoryId))) {
            const postElement = document.createElement('div');
            postElement.innerHTML = uiComponents.renderPost(post);
            feed.prepend(postElement);
            this.initReactionHandlers(postElement);
            this.initCommentHandlers(postElement);
            console.log('Appended WebSocket post:', post);
        }
    },

    initSignupForm() {
        document.body.innerHTML = uiComponents.renderSignup();
        const form = document.getElementById('signup-form');
        if (!form) return;
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const errorDiv = document.getElementById('error-info');
            const formData = {
                nickname: document.getElementById('nickname').value,
                first_name: document.getElementById('first_name').value,
                last_name: document.getElementById('last_name').value,
                email: document.getElementById('email').value,
                password: document.getElementById('password').value,
                age: parseInt(document.getElementById('age').value),
                gender: document.getElementById('gender').value,
            };
            try {
                await authService.signup(formData);
                errorDiv.textContent = '';
                this.showSuccessMessage('Signup successful! Please login.');
                window.location.hash = '#login';
            } catch (error) {
                errorDiv.textContent = error.message || 'Signup failed. Please try again.';
            }
        });
    },

    initLoginForm() {
        document.body.innerHTML = uiComponents.renderLogin();
        const form = document.getElementById('login-form');
        if (!form) return;
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const errorDiv = document.getElementById('error-info');
            try {
                await authService.login({
                    identifier: document.getElementById('username').value,
                    password: document.getElementById('password').value,
                });
                errorDiv.textContent = '';
                window.location.hash = '#home';
            } catch (error) {
                errorDiv.textContent = error.message || 'Login failed. Please check your credentials.';
            }
        });
    },

    async initHome() {
        try {
            const [categories, user, messages] = await Promise.all([
                postService.getCategories(),
                authService.getCurrentUser(),
                postService.getMessages(),
            ]);
            console.log('Fetched categories:', categories);
            document.body.innerHTML = uiComponents.renderHome(categories, user, messages);
            await this.renderPosts('0');
            this.initHomeHandlers();
            this.initMessageHandlers();
        } catch (error) {
            this.showErrorMessage('Failed to load home page: ' + (error.message || 'Unknown error'));
            console.error('Home initialization error:', error);
        }
    },

    async initProfile() {
        try {
            const [user, createdPosts, likedPosts] = await Promise.all([
                authService.getCurrentUser(),
                postService.getPostsByUserID(authService.getCurrentUserId()),
                postService.getLikedPosts(),
            ]);
            document.body.innerHTML = uiComponents.renderProfilePage(user, createdPosts, likedPosts);
            this.initProfileHandlers();
        } catch (error) {
            this.showErrorMessage('Failed to load profile page: ' + (error.message || 'Unknown error'));
            console.error('Profile initialization error:', error);
            document.body.innerHTML = uiComponents.renderProfilePage({}, [], []);
        }
    },

    initPostForm() {
        const form = document.getElementById('create-post-form');
        const postFormContainer = document.getElementById('post-form-container');
        if (!form || !postFormContainer) {
            console.error('Post form or container not found');
            return;
        }
        const imageInput = document.getElementById('post-image');
        const previewDiv = document.getElementById('image-preview');
        const previewImage = document.getElementById('preview-image');
        const removeImageBtn = document.getElementById('remove-image-btn');

        if (imageInput && previewDiv && previewImage) {
            imageInput.addEventListener('change', (e) => {
                const file = e.target.files[0];
                if (file) {
                    const reader = new FileReader();
                    reader.onload = (event) => {
                        previewImage.src = event.target.result;
                        previewDiv.classList.remove('hidden');
                    };
                    reader.readAsDataURL(file);
                }
            });
        }

        if (removeImageBtn && previewImage && previewDiv) {
            removeImageBtn.addEventListener('click', () => {
                previewImage.src = '';
                imageInput.value = '';
                previewDiv.classList.add('hidden');
            });
        }

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData();
            formData.append('title', document.getElementById('post-title').value);
            formData.append('content', document.getElementById('post-content').value);
            formData.append('category_id', document.getElementById('post-category').value);
            if (imageInput?.files[0]) {
                formData.append('image', imageInput.files[0]);
            }
            try {
                await postService.createPost(formData);
                form.reset();
                if (previewImage) previewImage.src = '';
                if (previewDiv) previewDiv.classList.add('hidden');
                this.showSuccessMessage('Post created successfully!');
                setTimeout(() => this.renderPosts(document.querySelector('.category-item.active')?.dataset.categoryId || '0'), 1000);
            } catch (error) {
                this.showErrorMessage(error.message || 'Failed to create post.');
            } finally {
                postFormContainer.classList.add('hidden');
            }
        });
    },

    initProfileHandlers() {
        const backBtn = document.getElementById('back-btn');
        if (backBtn) {
            backBtn.addEventListener('click', () => {
                window.location.hash = '#home';
            });
        }
        const logoutBtn = document.getElementById('logout-btn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', () => {
                authService.logout();
                window.location.hash = '#login';
            });
        }
        const profileIcon = document.querySelector('.profile-icon');
        if (profileIcon) {
            profileIcon.addEventListener('click', () => {
                window.location.hash = '#profile';
            });
        }
        const feed = document.getElementById('feed');
        if (feed) {
            this.initReactionHandlers(feed);
            this.initCommentHandlers(feed);
        }
    },

    initMessageHandlers() {
        const messageItems = document.querySelectorAll('.message-item');
        messageItems.forEach(item => {
            item.addEventListener('click', async () => {
                messageItems.forEach(i => i.classList.remove('active'));
                item.classList.add('active');
                const messageId = item.dataset.messageId;
                try {
                    const message = await postService.getMessage(messageId);
                    const messageView = document.createElement('div');
                    messageView.innerHTML = uiComponents.renderMessageView(message);
                    document.body.appendChild(messageView);
                    messageView.querySelector('.close-message-btn').addEventListener('click', () => {
                        messageView.remove();
                    });
                } catch (error) {
                    this.showErrorMessage('Failed to load message: ' + (error.message || 'Unknown error'));
                }
            });
        });
    },

    showSuccessMessage(message) {
        let msgDiv = document.getElementById('success-message');
        if (!msgDiv) {
            msgDiv = document.createElement('div');
            msgDiv.id = 'success-message';
            msgDiv.className = 'success-message';
            document.body.appendChild(msgDiv);
        }
        msgDiv.textContent = message;
        msgDiv.style.display = 'block';
        setTimeout(() => { msgDiv.style.display = 'none'; }, 3000);
    },

    showErrorMessage(message) {
        let msgDiv = document.getElementById('error-message');
        if (!msgDiv) {
            msgDiv = document.createElement('div');
            msgDiv.id = 'error-message';
            msgDiv.className = 'error-message';
            document.body.appendChild(msgDiv);
        }
        msgDiv.textContent = message;
        msgDiv.style.display = 'block';
        setTimeout(() => { msgDiv.style.display = 'none'; }, 3000);
    },

    initHomeHandlers() {
        const toggleBtn = document.getElementById('toggle-post-form-btn');
        const postFormContainer = document.getElementById('post-form-container');
        if (toggleBtn && postFormContainer) {
            toggleBtn.addEventListener('click', () => {
                postFormContainer.classList.toggle('hidden');
                if (!postFormContainer.classList.contains('hidden')) {
                    this.initPostForm();
                }
            });
        }
        const logoutBtn = document.getElementById('logout-btn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', () => {
                authService.logout();
                window.location.hash = '#login';
            });
        }
        const profileIcon = document.querySelector('.profile-icon');
        if (profileIcon) {
            profileIcon.addEventListener('click', () => {
                window.location.hash = '#profile';
            });
        }
        const categoryItems = document.querySelectorAll('.category-item');
        categoryItems.forEach(item => {
            item.addEventListener('click', async () => {
                categoryItems.forEach(i => i.classList.remove('active'));
                item.classList.add('active');
                const categoryId = item.dataset.categoryId;
                await this.renderPosts(categoryId);
            });
        });
    },

    async renderPosts(categoryId) {
        const feed = document.getElementById('feed');
        if (!feed) return;
        try {
            const posts = await postService.getPosts(categoryId);
            feed.innerHTML = uiComponents.renderFeed(posts);
            this.initReactionHandlers(feed);
            this.initCommentHandlers(feed);
        } catch (error) {
            this.showErrorMessage('Error loading posts: ' + (error.message || 'Unknown error'));
            console.error('Render posts error:', error);
            feed.innerHTML = uiComponents.renderFeed([]);
        }
    },

    initReactionHandlers(container = document) {
        container.querySelectorAll('.like-btn').forEach(button => {
            button.addEventListener('click', async () => {
                const postId = button.dataset.postId;
                try {
                    await postService.handleReaction(postId, 'like');
                    await this.renderPosts(document.querySelector('.category-item.active')?.dataset.categoryId || '0');
                } catch (error) {
                    this.showErrorMessage('Failed to like post: ' + (error.message || 'Unknown error'));
                }
            });
        });
        container.querySelectorAll('.dislike-btn').forEach(button => {
            button.addEventListener('click', async () => {
                const postId = button.dataset.postId;
                try {
                    await postService.handleReaction(postId, 'dislike');
                    await this.renderPosts(document.querySelector('.category-item.active')?.dataset.categoryId || '0');
                } catch (error) {
                    this.showErrorMessage('Failed to dislike post: ' + (error.message || 'Unknown error'));
                }
            });
        });
        container.querySelectorAll('.comment-btn').forEach(button => {
            button.addEventListener('click', () => {
                const postId = button.dataset.postId;
                const commentsSection = document.querySelector(`.post-card[data-post-id="${postId}"] .comments-section`);
                const commentForm = document.querySelector(`.post-card[data-post-id="${postId}"] .comment-form`);
                if (commentsSection && commentForm) {
                    commentsSection.classList.toggle('hidden');
                    if (!commentsSection.classList.contains('hidden')) {
                        commentForm.classList.remove('hidden');
                    }
                }
            });
        });
        container.querySelectorAll('.show-comment-form-btn').forEach(button => {
            button.addEventListener('click', () => {
                const postId = button.dataset.postId;
                const commentForm = document.querySelector(`.post-card[data-post-id="${postId}"] .comment-form`);
                if (commentForm) {
                    commentForm.classList.toggle('hidden');
                }
            });
        });
    },

    initCommentHandlers(container = document) {
        container.querySelectorAll('.comment-form').forEach(form => {
            form.addEventListener('submit', async (e) => {
                e.preventDefault();
                const postId = parseInt(form.dataset.postId);
                const content = form.querySelector('input[name="content"]').value;
                try {
                    await postService.addComment(postId, content);
                    form.querySelector('input[name="content"]').value = '';
                    form.classList.add('hidden');
                    this.showSuccessMessage('Comment added successfully!');
                    await this.renderPosts(document.querySelector('.category-item.active')?.dataset.categoryId || '0');
                } catch (error) {
                    this.showErrorMessage('Failed to add comment: ' + (error.message || 'Unknown error'));
                }
            });
        });
        container.querySelectorAll('.edit-comment-btn').forEach(button => {
            button.addEventListener('click', () => {
                const commentId = button.closest('[data-comment-id]').dataset.commentId;
                const commentForm = document.querySelector(`.edit-comment-form[data-comment-id="${commentId}"]`);
                if (commentForm) {
                    commentForm.classList.toggle('hidden');
                }
            });
        });
        container.querySelectorAll('.delete-comment-btn').forEach(button => {
            button.addEventListener('click', async () => {
                const commentId = button.closest('[data-comment-id]').dataset.commentId;
                try {
                    await postService.deleteComment(commentId);
                    this.showSuccessMessage('Comment deleted successfully!');
                    await this.renderPosts(document.querySelector('.category-item.active')?.dataset.categoryId || '0');
                } catch (error) {
                    this.showErrorMessage('Failed to delete comment: ' + (error.message || 'Unknown error'));
                }
            });
        });
        container.querySelectorAll('.edit-comment-form').forEach(form => {
            form.addEventListener('submit', async (e) => {
                e.preventDefault();
                const commentId = form.dataset.commentId;
                const content = form.querySelector('input[name="content"]').value;
                try {
                    await postService.editComment(commentId, content);
                    form.classList.add('hidden');
                    this.showSuccessMessage('Comment updated successfully!');
                    await this.renderPosts(document.querySelector('.category-item.active')?.dataset.categoryId || '0');
                } catch (error) {
                    this.showErrorMessage('Failed to update comment: ' + (error.message || 'Unknown error'));
                }
            });
            form.querySelector('.cancel-edit-btn').addEventListener('click', () => {
                form.classList.add('hidden');
            });
        });
    },
};

export default formHandlers;
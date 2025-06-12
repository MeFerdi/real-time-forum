import authService from './authService.js';
import postService from './postService.js';
import { uiComponents } from './uiComponents.js';
const formHandlers = {
    initSignupForm() {
        const form = document.getElementById('signup-form');
        if (!form) {
            console.warn('Signup form not found in DOM');
            return;
        }
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
                const response = await authService.signup(formData);
                errorDiv.textContent = '';
                window.location.hash = '#login';
            } catch (error) {
                errorDiv.textContent = error.message || 'Signup failed';
            }
        });
    },

    initLoginForm() {
        const form = document.getElementById('login-form');
        if (!form) {
            console.warn('Login form not found in DOM');
            return;
        }
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const errorDiv = document.getElementById('error-info');
            const formData = {
                identifier: document.getElementById('username').value,
                password: document.getElementById('password').value,
            };
            try {
                const response = await authService.login(formData);
                errorDiv.textContent = '';
                window.location.hash = '#home';
            } catch (error) {
                errorDiv.textContent = error.message || 'Login failed';
            }
        });
    },

    initPostForm() {
        const form = document.getElementById('create-post-form');
        if (!form) {
            console.warn('Post form not found in DOM');
            return;
        }
        const imageInput = document.getElementById('post-image');
        const previewDiv = document.getElementById('image-preview');
        const previewImage = document.getElementById('preview-image');
        const removeImageBtn = document.getElementById('remove-image-btn');
        const postFormContainer = document.getElementById('post-form-container');

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

        removeImageBtn.addEventListener('click', () => {
            previewImage.src = '';
            imageInput.value = '';
            previewDiv.classList.add('hidden');
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData();
            formData.append('title', document.getElementById('post-title').value);
            formData.append('content', document.getElementById('post-content').value);
            formData.append('category_id', document.getElementById('post-category').value);
            if (imageInput.files[0]) {
                formData.append('image', imageInput.files[0]);
            }
            try {
                const response = await postService.createPost(formData);
                form.reset();
                previewImage.src = '';
                previewDiv.classList.add('hidden');
                postFormContainer.classList.add('hidden');
                await this.renderPosts(0);
            } catch (error) {
                console.error('Post creation failed:', error);
            }
        });
    },

    initHomeHandlers() {
        const toggleBtn = document.getElementById('toggle-post-form-btn');
        const postFormContainer = document.getElementById('post-form-container');
        if (!toggleBtn || !postFormContainer) {
            console.warn('Toggle button or post form container not found');
            return;
        }
        toggleBtn.addEventListener('click', () => {
            postFormContainer.classList.toggle('hidden');
            if (!postFormContainer.classList.contains('hidden')) {
                this.initPostForm();
            }
        });
        const logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', async () => {
            await authService.logout();
            window.location.hash = '#login';
        });
    }
    const profileBtn = document.getElementById('profile-btn');
const profileDropdown = document.getElementById('profile-dropdown');
if (profileBtn && profileDropdown) {
    profileBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        profileDropdown.classList.toggle('hidden');
    });
    document.addEventListener('click', (e) => {
        if (!profileDropdown.classList.contains('hidden')) {
            profileDropdown.classList.add('hidden');
        }
    });
    profileDropdown.addEventListener('click', (e) => e.stopPropagation());
}

        const categoryItems = document.querySelectorAll('.category-item');
        console.log('Found category items:', categoryItems.length);
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
            feed.innerHTML = posts
                .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
                .map(post => uiComponents.renderPost(post)).join('');
            this.initReactionHandlers();
            this.initCommentHandlers();
        } catch (error) {
            console.error('Error rendering posts:', error);
        }
    },

    initReactionHandlers() {
        const likeButtons = document.querySelectorAll('.like-btn');
        const dislikeButtons = document.querySelectorAll('.dislike-btn');
        likeButtons.forEach(button => {
            button.addEventListener('click', async () => {
                const postId = button.dataset.postId;
                try {
                    await postService.handleReaction(postId, 'like');
                    await this.renderPosts(document.querySelector('.category-item.active').dataset.categoryId);
                } catch (error) {
                    console.error('Error liking post:', error);
                }
            });
        });
        dislikeButtons.forEach(button => {
            button.addEventListener('click', async () => {
                const postId = button.dataset.postId;
                try {
                    await postService.handleReaction(postId, 'dislike');
                    await this.renderPosts(document.querySelector('.category-item.active').dataset.categoryId);
                } catch (error) {
                    console.error('Error disliking post:', error);
                }
            });
        });

        const commentButtons = document.querySelectorAll('.comment-btn');
        commentButtons.forEach(button => {
            button.addEventListener('click', () => {
                const postId = button.dataset.postId;
                const commentsSection = document.querySelector(`.post-card[data-post-id="${postId}"] .comments-section`);
                commentsSection.classList.toggle('hidden');
            });
        });
    },

    initCommentHandlers() {
        const commentForms = document.querySelectorAll('.comment-form');
        commentForms.forEach(form => {
            form.addEventListener('submit', async (e) => {
                e.preventDefault();
                const postId = parseInt(form.dataset.postId);
                const content = form.querySelector('input[name="content"]').value;
                try {
                    await postService.addComment(postId, { content });
                    form.reset();
                    await this.renderPosts(document.querySelector('.category-item.active').dataset.categoryId);
                } catch (error) {
                    console.error('Error creating comment:', error);
                }
            });
        });
    },
};

export default formHandlers;
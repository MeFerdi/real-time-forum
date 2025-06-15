class Views {
    constructor() {
        this.bindEvents();
        this.currentUser = null;
        this.loadCategories();
    }

    bindEvents() {
        // Navigation events
        document.getElementById('loginBtn').addEventListener('click', () => router.navigate('/login'));
        document.getElementById('registerBtn').addEventListener('click', () => router.navigate('/register'));
        document.getElementById('logoutBtn').addEventListener('click', () => this.handleLogout());

        // Form submissions
        document.getElementById('login-form').addEventListener('submit', (e) => this.handleLogin(e));
        document.getElementById('register-form').addEventListener('submit', (e) => this.handleRegister(e));
        document.getElementById('chat-form')?.addEventListener('submit', (e) => this.handleChatMessage(e));

        // Feed events
        document.getElementById('categoryFilter').addEventListener('change', (e) => this.loadPosts(e.target.value));
        document.getElementById('newPostBtn').addEventListener('click', () => this.showNewPostModal());
    }

    // Authentication handlers
    async handleLogin(e) {
        e.preventDefault();
        const formData = new FormData(e.target);
        const credentials = {
            login: formData.get('login'),
            password: formData.get('password')
        };

        const result = await API.login(credentials);
        if (result.success) {
            this.currentUser = result.data;
            router.setAuthenticated(true);
            router.navigate('/');
            this.loadPosts();
        } else {
            alert('Login failed: ' + result.error);
        }
    }

    async handleRegister(e) {
        e.preventDefault();
        const formData = new FormData(e.target);
        const userData = {
            username: formData.get('username'),
            email: formData.get('email'),
            password: formData.get('password'),
            first_name: formData.get('first_name'),
            last_name: formData.get('last_name'),
            age: parseInt(formData.get('age')),
            gender: formData.get('gender')
        };

        const result = await API.register(userData);
        if (result.success) {
            router.navigate('/login');
            alert('Registration successful! Please log in.');
        } else {
            alert('Registration failed: ' + result.error);
        }
    }

    async handleLogout() {
        const result = await API.logout();
        if (result.success) {
            this.currentUser = null;
            router.setAuthenticated(false);
            router.navigate('/login');
        }
    }

    // Post and comment handlers
    async loadPosts(category = '') {
        const result = await API.getPosts(category);
        if (result.success) {
            const container = document.getElementById('posts-container');
            container.innerHTML = result.data.map(post => this.renderPost(post)).join('');
            
            // Add click handlers to posts
            container.querySelectorAll('.post-card').forEach(card => {
                card.addEventListener('click', () => this.loadPost(card.dataset.postId));
            });
        }
    }

    async loadPost(postId) {
        const result = await API.getPost(postId);
        if (result.success) {
            this.currentPostId = result.data.id; // Set current post ID
            document.getElementById('post-detail').innerHTML = this.renderPostDetail(result.data);
            
            // Rebind comment form submit event after rendering
            const commentForm = document.getElementById('comment-form');
            if (commentForm) {
                commentForm.addEventListener('submit', (e) => this.handleComment(e));
            }
            
            router.navigate('/post');
        }
    }

    async handleComment(e) {
        e.preventDefault();
        const formData = new FormData(e.target);
        const commentData = {
            post_id: this.currentPostId,
            content: formData.get('content')
        };

        const result = await API.createComment(commentData);
        if (result.success) {
            e.target.reset();
            const commentsContainer = document.getElementById('comments-container');
            const newComment = result.data;
            
            // Add the new comment to the UI
            const commentElement = document.createElement('div');
            commentElement.className = 'comment';
            commentElement.innerHTML = `
                <p>${newComment.content}</p>
                <div class="comment-meta">
                    <span>By ${newComment.author.username}</span>
                    <span>${new Date(newComment.created_at).toLocaleString()}</span>
                    <button onclick="views.handleCommentLike(${newComment.id})">
                        💗 ${newComment.like_count || 0}
                    </button>
                </div>
            `;
            commentsContainer.insertBefore(commentElement, commentsContainer.firstChild);
        } else {
            alert('Failed to add comment: ' + (result.error || 'Unknown error'));
        }
    }

    showNewPostModal() {
        const modal = document.getElementById('new-post-modal');
        const closeBtn = modal.querySelector('.close');
        const form = document.getElementById('new-post-form');
        
        modal.classList.add('active');
        
        closeBtn.onclick = () => modal.classList.remove('active');
        window.onclick = (e) => {
            if (e.target === modal) {
                modal.classList.remove('active');
            }
        };

        // Handle form submission
        form.onsubmit = async (e) => {
            e.preventDefault();
            const formData = new FormData(form);
            const postData = {
                title: formData.get('title'),
                content: formData.get('content'),
                category_ids: [parseInt(formData.get('category'))]
            };

            const result = await API.createPost(postData);
            if (result.success) {
                modal.classList.remove('active');
                form.reset();
                this.loadPosts();
            } else {
                alert('Failed to create post: ' + result.error);
            }
        };
    }

    async loadCategories() {
        const result = await API.getCategories();
        if (result.success) {
            const categorySelects = document.querySelectorAll('select[name="category"]');
            const options = result.data.map(category => 
                `<option value="${category.id}">${category.name}</option>`
            ).join('');
            
            categorySelects.forEach(select => {
                select.innerHTML = '<option value="">Select Category</option>' + options;
            });
        }
    }

    // UI Rendering methods
    renderPost(post) {
        return `
            <div class="post-card" data-post-id="${post.id}">
                <h3>${post.title}</h3>
                <p>${post.content.substring(0, 150)}...</p>
                <div class="post-meta">
                    <span>By ${post.author.username}</span>
                    <span>💗 ${post.like_count}</span>
                </div>
            </div>
        `;
    }

    renderPostDetail(post) {
        this.currentPostId = post.id;
        return `
            <div class="post-full">
                <h2>${post.title}</h2>
                <p>${post.content}</p>
                <div class="post-meta">
                    <span>By ${post.author.username}</span>
                    <button onclick="views.handleLike(${post.id})">
                        💗 ${post.like_count}
                    </button>
                </div>
                <div id="comments-section">
                    <h3>Comments</h3>
                    <form id="comment-form" class="comment-form">
                        <textarea name="content" placeholder="Write a comment..." required></textarea>
                        <button type="submit">Add Comment</button>
                    </form>
                    <div id="comments-container">
                        ${this.renderComments(post.comments || [])}
                    </div>
                </div>
            </div>
        `;
    }

    renderComments(comments) {
        return comments.map(comment => `
            <div class="comment">
                <p>${comment.content}</p>
                <div class="comment-meta">
                    <span>By ${comment.author.username}</span>
                    <span>${new Date(comment.created_at).toLocaleString()}</span>
                    <button onclick="views.handleCommentLike(${comment.id})">
                        💗 ${comment.like_count}
                    </button>
                </div>
            </div>
        `).join('');
    }

    async handleLike(postId) {
        const result = await API.likePost(postId);
        if (result.success) {
            this.loadPost(postId);
        }
    }

    // Initialize the views
    async init() {
        try {
            const profile = await API.getProfile();
            if (profile.success) {
                this.currentUser = profile.data;
                router.setAuthenticated(true);
                this.loadPosts();
            }
        } catch (error) {
            console.error('Failed to load profile:', error);
            router.setAuthenticated(false);
        }
    }
}

const views = new Views();
import API from './api.js'
import router from './router.js'
class Views {
    constructor() {
        this.api = API;
        this.router = router;
        this.currentUser = null;
        this.categoriesLoaded = false;
        this.chatHistory = new Map();
        this.activeChat = null;
        this.messageCallbacks = [];
        this.bindEvents();
    }

    formatRelativeTime(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const seconds = Math.floor((now - date) / 1000);
        
        if (seconds < 60) {
            return 'just now';
        }
        
        const minutes = Math.floor(seconds / 60);
        if (minutes < 60) {
            return `${minutes} ${minutes === 1 ? 'minute' : 'minutes'} ago`;
        }
        
        const hours = Math.floor(minutes / 60);
        if (hours < 24) {
            return `${hours} ${hours === 1 ? 'hour' : 'hours'} ago`;
        }
        
        const days = Math.floor(hours / 24);
        if (days < 7) {
            return `${days} ${days === 1 ? 'day' : 'days'} ago`;
        }
        
        const weeks = Math.floor(days / 7);
        if (weeks < 4) {
            return `${weeks} ${weeks === 1 ? 'week' : 'weeks'} ago`;
        }
        
        const months = Math.floor(days / 30);
        if (months < 12) {
            return `${months} ${months === 1 ? 'month' : 'months'} ago`;
        }
        
        const years = Math.floor(days / 365);
        return `${years} ${years === 1 ? 'year' : 'years'} ago`;
    }
    getCurrentUser() {
        return this.currentUser;
    }

    bindEvents() {
        // Navigation events
        document.getElementById('loginBtn').addEventListener('click', () => router.navigate('/login'));
        document.getElementById('registerBtn').addEventListener('click', () => router.navigate('/register'));
        document.getElementById('logoutBtn').addEventListener('click', () => this.handleLogout());
        document.getElementById('homeBtn').addEventListener('click', () => router.navigate('/'));
        document.getElementById('chatBtn').addEventListener('click', () => router.navigate('/chat'));
        document.getElementById('profileBtn').addEventListener('click', () => this.showProfile());
        document.querySelector('.nav-left h1').addEventListener('click', () => router.navigate('/'));

        // Form submissions
        document.getElementById('login-form').addEventListener('submit', (e) => this.handleLogin(e));
        document.getElementById('register-form').addEventListener('submit', (e) => this.handleRegister(e));
        document.getElementById('chat-form')?.addEventListener('submit', (e) => this.handleChatMessage(e));

        // Feed events
        document.getElementById('categoryFilter').addEventListener('change', (e) => this.loadPosts(e.target.value));
        document.getElementById('newPostBtn').addEventListener('click', () => this.toggleQuickPostForm());
        document.getElementById('quick-post-form').addEventListener('submit', (e) => this.handleQuickPost(e));

        // Mobile menu events
        document.getElementById('hamburger-menu').addEventListener('click', () => this.toggleMobileMenu());
        document.getElementById('homeBtn-mobile').addEventListener('click', () => router.navigate('/'));
        document.getElementById('chatBtn-mobile').addEventListener('click', () => router.navigate('/chat'));
        document.getElementById('loginBtn-mobile').addEventListener('click', () => router.navigate('/login'));
        document.getElementById('registerBtn-mobile').addEventListener('click', () => router.navigate('/register'));
        document.getElementById('profileBtn-mobile').addEventListener('click', () => this.showProfile());
    }

    handleChatMessage(e) {
        e.preventDefault();
        const form = e.target;
        const content = form.message.value.trim();
        const receiverId = form.getAttribute('data-receiver-id');
        
        if (content && receiverId) {
            this.api.sendMessage(parseInt(receiverId), content);
            form.reset();
        }
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
            this.categoriesLoaded = false; // Reset categories flag
            router.navigate('/');
            await this.loadCategories(); // Load categories first
            this.updateProfileCard(); // Update profile card with user info
            this.loadPosts(); // Then load posts
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
            
            // Close any open profile modal first
            const modal = document.querySelector('.modal.active');
            if (modal) {
                document.body.removeChild(modal);
            }
            
            router.navigate('/login');
            // Add a success message to the login form
            const loginSection = document.getElementById('login-section');
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message success';
            messageDiv.textContent = 'Successfully logged out. Please log in to continue.';
            const existingMessage = loginSection.querySelector('.message');
            if (existingMessage) {
                loginSection.removeChild(existingMessage);
            }
            loginSection.insertBefore(messageDiv, document.getElementById('login-form'));
        } else {
            alert('Logout failed: ' + (result.error || 'Unknown error'));
        }
    }
    async initializeChat() {
        const users = await API.getUsersWithChats();
        if (users.success) {
            this.renderUserList(users.data);
        }
        this.setupChatHandlers();
    }

    async loadPosts(category = '', filter = '') {
        try {
            const container = document.getElementById('posts-container');
            if (!container) return;
    
            // Show loading state
            container.innerHTML = '<div class="loading">Loading posts...</div>';
    
            const result = await this.api.getPosts(category);
            if (!result || !result.success) {
                container.innerHTML = `
                    <div class="no-posts-message">
                        <p>Failed to load posts</p>
                    </div>
                `;
                return;
            }
    
            let filteredPosts = result.data || [];
            let filterMessage = '';
    
            // Validate current user before filtering
            if (this.currentUser) {
                // Apply filters if data exists
                if (filter === 'my-posts') {
                    filteredPosts = filteredPosts.filter(post => 
                        post.author.username === this.currentUser.username
                    );
                    filterMessage = 'You haven\'t created any posts yet.';
                } else if (filter === 'liked-posts') {
                    filteredPosts = filteredPosts.filter(post => 
                        post.like_count > 0
                    );
                    filterMessage = 'You haven\'t liked any posts yet.';
                }
            }
    
            // Update UI based on results
            if (!filteredPosts.length) {
                container.innerHTML = `
                    <div class="no-posts-message">
                        <p>${filterMessage || 'No posts found.'}</p>
                        ${filter ? `<button onclick="views.loadPosts()" class="action-btn">View All Posts</button>` : ''}
                    </div>
                `;
                return;
            }
    
            // Render posts and add event listeners
            container.innerHTML = filteredPosts.map(post => this.renderPost(post)).join('');
            container.querySelectorAll('.post-card').forEach(card => {
                card.addEventListener('click', () => this.loadPost(card.dataset.postId));
            });
    
            // Update profile if needed
            if (this.currentUser) {
                this.updateProfileCard();
            }
    
        } catch (error) {
            console.error('Error loading posts:', error);
            const container = document.getElementById('posts-container');
            if (container) {
                container.innerHTML = `
                    <div class="no-posts-message error">
                        <p>Error loading posts. Please try again.</p>
                    </div>
                `;
            }
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

    toggleQuickPostForm() {
        const form = document.getElementById('quick-post-form');
        const btn = document.getElementById('newPostBtn');
        form.classList.toggle('hidden');
        
        if (!form.classList.contains('hidden')) {
            form.querySelector('textarea').focus();
            btn.style.display = 'none';
        } else {
            btn.style.display = 'block';
        }
    }

    async handleQuickPost(e) {
        e.preventDefault();
        const form = e.target;
        const formData = new FormData(form);
        
        const postData = {
            title: formData.get('title'),
            content: formData.get('content'),
            category_ids: [parseInt(formData.get('category'))]
        };

        const result = await API.createPost(postData);
        if (result.success) {
            form.reset();
            this.toggleQuickPostForm();
            this.loadPosts();
        } else {
            alert('Failed to create post: ' + result.error);
        }
    }

    async loadCategories() {
        // If categories are already loaded and dropdowns exist, don't reload
        const categoryFilter = document.getElementById('categoryFilter');
        const quickPostCategory = document.querySelector('#quick-post-form select[name="category"]');
        
        if (this.categoriesLoaded && categoryFilter?.children.length > 0) {
            return;
        }

        const result = await API.getCategories();
        if (result.success) {
            // Create options array from the categories, sort them by name
            const sortedCategories = [...result.data].sort((a, b) => a.name.localeCompare(b.name));
            
            // Reset the loaded flag if selects don't exist
            if (!categoryFilter || !quickPostCategory) {
                this.categoriesLoaded = false;
                return;
            }

            // Function to populate a select element
            const populateSelect = (select, defaultText) => {
                select.innerHTML = ''; // Clear existing options
                select.appendChild(new Option(defaultText, '')); // Add default option
                sortedCategories.forEach(category => {
                    select.appendChild(new Option(category.name, category.id));
                });
            };

            // Update both dropdowns
            if (categoryFilter) {
                populateSelect(categoryFilter, 'All Categories');
            }
            if (quickPostCategory) {
                populateSelect(quickPostCategory, 'Select Category');
            }

            this.categoriesLoaded = true;
        }
    }
    async loadUsers() {
        try {
            console.log('Starting user load...');
            
            // Get all registered users
            const allUsers = await this.api.getUsers();
            console.log('All users response:', allUsers);
            
            if (allUsers.success) {
                // Store all users in map
                allUsers.data.forEach(user => {
                    this.users.set(user.id, { ...user, online: false });
                });
                console.log(`Stored ${this.users.size} users in map`);
        
                // Get online users
                const onlineUsers = await this.api.getOnlineUsers();
                console.log('Online users response:', onlineUsers);
                
                if (onlineUsers.success) {
                    // Update online status
                    let onlineCount = 0;
                    onlineUsers.data.forEach(user => {
                        if (this.users.has(user.id)) {
                            const userData = this.users.get(user.id);
                            this.users.set(user.id, { ...userData, online: true });
                            onlineCount++;
                        }
                    });
                    console.log(`Updated ${onlineCount} users to online status`);
                }
        
                // Render complete list
                const usersList = Array.from(this.users.values());
                console.log(`Rendering ${usersList.length} users`);
                this.renderUserList(usersList);
            }
        } catch (error) {
            console.error('Failed to load users:', error);
            console.error('Error stack:', error.stack);
        }
    }

    renderUserList(users) {
        const usersList = document.getElementById('users-list');
        if (!usersList) return;

        usersList.innerHTML = users.map(user => `
            <div class="user-item ${user.online ? 'online' : ''}" data-user-id="${user.id}">
                <div class="user-avatar">
                    <img src="https://ui-avatars.com/api/?name=${user.username}" alt="${user.username}">
                </div>
                <div class="user-info">
                    <span class="user-name">${user.username}</span>
                    <span class="last-message">${user.last_message || ''}</span>
                </div>
                <span class="status-indicator"></span>
            </div>
        `).join('');

        // Add click handlers
        usersList.querySelectorAll('.user-item').forEach(item => {
            item.addEventListener('click', () => this.loadChat(item.dataset.userId));
        });
    }
    updateUserStatus(userId, online) {
        if (online) {
            this.onlineUsers.add(userId);
        } else {
            this.onlineUsers.delete(userId);
        }
        this.renderUserList();
    }

    async loadChat(userId) {
        this.activeChat = userId;
        const chatContainer = document.getElementById('chat-messages');
        const messages = await API.getChatHistory(userId);
        
        if (messages.success) {
            this.renderChatMessages(messages.data);
            this.scrollToBottom();
        }

        // Update UI to show active chat
        document.querySelectorAll('.user-item').forEach(item => {
            item.classList.toggle('active', item.dataset.userId === userId);
        });
    }

    renderChatMessages(messages) {
        const chatContainer = document.getElementById('chat-messages');
        if (!chatContainer) return;

        chatContainer.innerHTML = messages.map(msg => `
            <div class="message ${msg.sender_id === this.currentUser.id ? 'sent' : 'received'}">
                <div class="message-content">${msg.content}</div>
                <div class="message-meta">
                    <span class="message-time">${this.formatRelativeTime(msg.created_at)}</span>
                </div>
            </div>
        `).join('');
    }

    setupChatHandlers() {
        const chatForm = document.getElementById('chat-form');
        if (chatForm) {
            chatForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                const input = chatForm.querySelector('input[name="message"]');
                const message = input.value.trim();
                
                if (message && this.activeChat) {
                    API.sendMessage(this.activeChat, message);
                    input.value = '';
                }
            });
        }

        // Add message listener
        API.onMessage(message => {
            if (message.type === 'chat') {
                this.handleIncomingMessage(message.data);
            }
        });
    }

    handleIncomingMessage(message) {
        if (message.sender_id === this.activeChat || 
            message.sender_id === this.currentUser.id) {
            this.appendMessage(message);
        }
        this.updateUserList();
    }

    appendMessage(message) {
        const chatContainer = document.getElementById('chat-messages');
        if (!chatContainer) return;

        const messageEl = document.createElement('div');
        messageEl.className = `message ${message.sender_id === this.currentUser.id ? 'sent' : 'received'}`;
        messageEl.innerHTML = `
            <div class="message-content">${message.content}</div>
            <div class="message-meta">
                <span class="message-time">${this.formatRelativeTime(message.created_at)}</span>
            </div>
        `;

        chatContainer.appendChild(messageEl);
        this.scrollToBottom();
    }

    scrollToBottom() {
        const chatContainer = document.getElementById('chat-messages');
        if (chatContainer) {
            chatContainer.scrollTop = chatContainer.scrollHeight;
        }
    }

    // UI Rendering methods
    renderPost(post) {
        return `
            <div class="post-card" data-post-id="${post.id}">
                <div class="post-header">
                    <div class="user-info">
                        <div class="avatar">
                            <img src="https://ui-avatars.com/api/?name=${post.author.username}&background=random" alt="${post.author.username}'s avatar" />
                        </div>
                        <div class="post-meta-info">
                            <span class="username">${post.author.username}</span>
                            <span class="timestamp">${this.formatRelativeTime(post.created_at)}</span>
                        </div>
                    </div>
                </div>
                <h3>${post.title}</h3>
                <p>${post.content.substring(0, 150)}...</p>
                <div class="post-meta">
                    <div class="post-actions">
                        <button onclick="event.stopPropagation(); views.handleLike(${post.id})" class="action-btn">
                            💗 ${post.like_count || 0}
                        </button>
                        <button onclick="event.stopPropagation(); views.loadPost(${post.id})" class="action-btn">
                            💬 ${post.comments ? post.comments.length : 0}
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    renderPostDetail(post) {
        this.currentPostId = post.id;
        return `
            <div class="post-full">
                <div class="post-header">
                    <div class="user-info">
                        <div class="avatar">
                            <img src="https://ui-avatars.com/api/?name=${post.author.username}&background=random" alt="${post.author.username}'s avatar" />
                        </div>
                        <div class="post-meta-info">
                            <span class="username">${post.author.username}</span>
                            <span class="timestamp">${this.formatRelativeTime(post.created_at)}</span>
                        </div>
                    </div>
                </div>
                <h2>${post.title}</h2>
                <p>${post.content}</p>
                <div class="post-meta">
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
                    <div class="user-info">
                        <div class="avatar">
                            <img src="https://ui-avatars.com/api/?name=${comment.author.username}&background=random" alt="${comment.author.username}'s avatar" />
                        </div>
                        <div class="post-meta-info">
                            <span class="username">${comment.author.username}</span>
                            <span class="timestamp">${this.formatRelativeTime(comment.created_at)}</span>
                        </div>
                    </div>
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
            // Update all like buttons for this post (both in cards and full view)
            const likeButtons = document.querySelectorAll(`button[onclick*="views.handleLike(${postId})"]`);
            likeButtons.forEach(button => {
                button.innerHTML = `💗 ${result.data.like_count}`;
            });
        } else {
            alert('Failed to like post: ' + (result.error || 'Unknown error'));
        }
    }

    async handleCommentLike(commentId) {
        const result = await API.likeComment(commentId);
        if (result.success) {
            // Update only the comment like count
            const likeButton = document.querySelector(`button[onclick="views.handleCommentLike(${commentId})"]`);
            if (likeButton) {
                likeButton.innerHTML = `💗 ${result.data.like_count}`;
            }
        } else {
            alert('Failed to like comment: ' + (result.error || 'Unknown error'));
        }
    }

    async showProfile() {
        try {
            const result = await API.getProfile();
            if (result.success) {
                const modal = document.createElement('div');
                modal.className = 'modal active';
                modal.innerHTML = `
                    <div class="modal-content">
                        <span class="close">&times;</span>
                        <h2>Profile</h2>
                        <div class="profile-info">
                            <p><strong>Username:</strong> ${result.data.username}</p>
                            <p><strong>Email:</strong> ${result.data.email}</p>
                            <p><strong>Name:</strong> ${result.data.first_name} ${result.data.last_name}</p>
                            <p><strong>Age:</strong> ${result.data.age}</p>
                            <p><strong>Gender:</strong> ${result.data.gender}</p>
                            <p><strong>Member since:</strong> ${new Date(result.data.created_at).toLocaleDateString()}</p>
                        </div>
                        <div class="profile-actions">
                            <button onclick="views.handleLogout()" class="signout-btn">Sign Out</button>
                        </div>
                    </div>
                `;
                
                document.body.appendChild(modal);
                
                const closeBtn = modal.querySelector('.close');
                closeBtn.onclick = () => document.body.removeChild(modal);
                window.onclick = (e) => {
                    if (e.target === modal) {
                        document.body.removeChild(modal);
                    }
                };
            }
        } catch (error) {
            console.error('Failed to load profile:', error);
            alert('Failed to load profile');
        }
    }

    updateProfileCard() {
        if (!this.currentUser) return;

        const profileName = document.getElementById('profile-name');
        const profileImage = document.getElementById('profile-image');
        const postCount = document.getElementById('post-count');
        
        if (profileName) {
            profileName.textContent = `${this.currentUser.first_name}`;
        }
        
        if (profileImage) {
            profileImage.className = 'default-avatar';
            profileImage.innerHTML = '<i class="fas fa-user"></i>';
        }

        // Update post count
        this.updateUserStats();
    }

    async updateUserStats(userData) {
        try {
            if (!userData) {
                console.warn('No user data available for stats update');
                return;
            }
    
            const postCount = document.getElementById('post-count');
            if (postCount) {
                postCount.textContent = userData.post_count || '0';
            }
        } catch (error) {
            console.error('Error updating user stats:', error);
        }
    }

    toggleMobileMenu() {
        const mobileMenu = document.getElementById('mobile-menu');
        mobileMenu.classList.toggle('active');
        
        // Close menu when clicking outside
        const closeMenu = (e) => {
            if (!e.target.closest('#mobile-menu') && !e.target.closest('#hamburger-menu')) {
                mobileMenu.classList.remove('active');
                document.removeEventListener('click', closeMenu);
            }
        };
        
        if (mobileMenu.classList.contains('active')) {
            setTimeout(() => {
                document.addEventListener('click', closeMenu);
            }, 0);
        }
    }

    // Initialize the views
    async init() {
        try {
            const profile = await API.getProfile();
            if (profile.success) {
                this.currentUser = profile.data;
                router.setAuthenticated(true);
                await this.loadUsers();
                await this.loadCategories(); // Load categories first
                this.updateProfileCard(); // Update profile card with user info
                this.loadPosts(); // Then load posts
                if (window.location.pathname === '/chat') {
                    this.initializeChat();
                }
            }
        } catch (error) {
            console.error('Failed to load profile:', error);
            router.setAuthenticated(false);
        }
    }
}
const views = new Views();
export default views;
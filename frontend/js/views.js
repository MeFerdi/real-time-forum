class Views {
    constructor() {
        this.bindEvents();
        this.currentUser = null;
        this.categoriesLoaded = false;
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

    bindEvents() {
        // Navigation events
        document.getElementById('loginBtn')?.addEventListener('click', () => router.navigate('/login'));
        document.getElementById('registerBtn')?.addEventListener('click', () => router.navigate('/register'));
        document.getElementById('homeBtn')?.addEventListener('click', () => router.navigate('/'));
        document.getElementById('chatBtn')?.addEventListener('click', () => router.navigate('/chat'));
        document.getElementById('profileBtn')?.addEventListener('click', () => this.showProfile());
        document.querySelector('.nav-left h1')?.addEventListener('click', () => router.navigate('/'));

        // Mobile navigation
        document.getElementById('homeBtn-mobile')?.addEventListener('click', () => router.navigate('/'));
        document.getElementById('chatBtn-mobile')?.addEventListener('click', () => router.navigate('/chat'));
        document.getElementById('profileBtn-mobile')?.addEventListener('click', () => this.showProfile());

        // Mobile menu toggle
        const hamburgerBtn = document.getElementById('hamburger-menu');
        const mobileMenu = document.getElementById('mobile-menu');

        hamburgerBtn?.addEventListener('click', () => {
            mobileMenu.classList.toggle('active');
        });

        // Close mobile menu when clicking outside
        document.addEventListener('click', (e) => {
            if (!hamburgerBtn?.contains(e.target) && !mobileMenu?.contains(e.target)) {
                mobileMenu?.classList.remove('active');
            }
        });

        // Form submissions
        document.getElementById('login-form')?.addEventListener('submit', (e) => this.handleLogin(e));
        document.getElementById('register-form')?.addEventListener('submit', (e) => this.handleRegister(e));
        // Chat form is now handled by chat.js

        // Feed events
        document.getElementById('categoryFilter')?.addEventListener('change', (e) => this.loadPosts(e.target.value));
        document.getElementById('newPostBtn')?.addEventListener('click', () => this.toggleQuickPostForm());
        document.getElementById('quick-post-form')?.addEventListener('submit', (e) => this.handleQuickPost(e));

        // Mobile menu events
        document.getElementById('hamburger-menu')?.addEventListener('click', () => this.toggleMobileMenu());
        document.getElementById('homeBtn-mobile')?.addEventListener('click', () => router.navigate('/'));
        document.getElementById('chatBtn-mobile')?.addEventListener('click', () => router.navigate('/chat'));
        document.getElementById('loginBtn-mobile')?.addEventListener('click', () => router.navigate('/login'));
        document.getElementById('registerBtn-mobile')?.addEventListener('click', () => router.navigate('/register'));
        document.getElementById('profileBtn-mobile')?.addEventListener('click', () => this.showProfile());
    }

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

        // --- Ensure chat is initialized for the new session ---
        if (window.Chat) {
            window.Chat.isInitialized = false; // Force re-initialization
            window.Chat.initializeChat();
        }
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

        // Disconnect WebSocket
        if (window.wsClient) {
            window.wsClient.disconnect();
        }

        // --- Reset chat state on logout ---
        if (window.Chat) {
            window.Chat.isInitialized = false;
            if (window.Chat.disconnect) window.Chat.disconnect();
        }

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

    // Post and comment handlers
    async loadPosts(category = '', filter = '') {
        // Check if user is authenticated before making API calls
        if (!this.currentUser) {
            console.log('User not authenticated, skipping loadPosts');
            return;
        }

        try {
            const result = await API.getPosts(category);
            console.log('loadPosts API response:', result); // Debug logging

            if (result.success && result.data && Array.isArray(result.data)) {
                let filteredPosts = result.data;
                let filterMessage = '';

                // Apply additional filters if specified
                if (filter === 'my-posts') {
                    filteredPosts = result.data.filter(post =>
                        post.author && post.author.username === this.currentUser.username
                    );
                    filterMessage = 'You haven\'t created any posts yet.';
                } else if (filter === 'liked-posts') {
                    filteredPosts = result.data.filter(post =>
                        post.like_count > 0
                    );
                    filterMessage = 'You haven\'t liked any posts yet.';
                } else if (category && category !== '') {
                    // If a specific category is selected, set appropriate message
                    filterMessage = 'No posts found in this category.';
                }

                const container = document.getElementById('posts-container');
                if (!container) {
                    console.error('Posts container not found');
                    return;
                }

                if (filteredPosts.length === 0) {
                    container.innerHTML = `
                        <div class="no-posts-message">
                            <p>${filterMessage || 'No posts found.'}</p>
                            ${(filter || category) ? `<button onclick="window.views.loadPosts()" class="action-btn">View All Posts</button>` : ''}
                        </div>
                    `;
                } else {
                    container.innerHTML = filteredPosts.map(post => this.renderPost(post)).join('');

                    // Add click handlers to posts
                    container.querySelectorAll('.post-card').forEach(card => {
                        card.addEventListener('click', () => this.loadPost(card.dataset.postId));
                    });
                }

                // Update the profile card with latest stats
                this.updateProfileCard();
            } else {
                console.warn('Invalid API response for posts:', result);
                // Show error message to user
                const container = document.getElementById('posts-container');
                if (container) {
                    container.innerHTML = `
                        <div class="no-posts-message">
                            <p>Failed to load posts. Please try again.</p>
                            <button onclick="window.views.loadPosts()" class="action-btn">Retry</button>
                        </div>
                    `;
                }
            }
        } catch (error) {
            console.error('Failed to load posts:', error);
            // Show error message to user
            const container = document.getElementById('posts-container');
            if (container) {
                container.innerHTML = `
                    <div class="no-posts-message">
                        <p>Failed to load posts. Please try again.</p>
                        <button onclick="window.views.loadPosts()" class="action-btn">Retry</button>
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
                    <div class="user-info">
                        <div class="avatar">
                            <img src="https://ui-avatars.com/api/?name=${newComment.author.username}&background=random" alt="${newComment.author.username}'s avatar" />
                        </div>
                        <div class="post-meta-info">
                            <span class="username">${newComment.author.username}</span>
                            <span class="timestamp">${this.formatRelativeTime(newComment.created_at)}</span>
                        </div>
                    </div>
                    <span class="action-icon like-icon" onclick="window.views.handleCommentLike(${newComment.id})" data-comment-id="${newComment.id}">
                        <i class="fas fa-heart"></i>
                        <span class="count">${newComment.like_count || 0}</span>
                    </span>
                </div>
            `;
            commentsContainer.insertBefore(commentElement, commentsContainer.firstChild);

            // Hide the comment form after successful submission
            e.target.classList.add('hidden');

            // Update comment count in the UI
            const commentCountSpans = document.querySelectorAll('.comment-icon .count');
            commentCountSpans.forEach(span => {
                const currentCount = parseInt(span.textContent) || 0;
                span.textContent = currentCount + 1;
            });

            // Update comments section header
            const commentsHeader = document.querySelector('#comments-section h3');
            if (commentsHeader) {
                const currentCount = parseInt(commentsHeader.textContent.match(/\d+/)?.[0] || '0');
                commentsHeader.textContent = `Comments (${currentCount + 1})`;
            }
        } else {
            alert('Failed to add comment: ' + (result.error || 'Unknown error'));
        }
    }

    toggleCommentForm() {
        const commentForm = document.getElementById('comment-form');
        if (commentForm) {
            commentForm.classList.toggle('hidden');
            if (!commentForm.classList.contains('hidden')) {
                const textarea = commentForm.querySelector('textarea');
                if (textarea) {
                    textarea.focus();
                }
            }
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
                        <span class="action-icon like-icon" onclick="event.stopPropagation(); window.views.handleLike(${post.id})" data-post-id="${post.id}">
                            <i class="fas fa-heart"></i>
                            <span class="count">${post.like_count || 0}</span>
                        </span>
                        <span class="action-icon comment-icon" onclick="event.stopPropagation(); window.views.loadPost(${post.id})">
                            <i class="fas fa-comment"></i>
                            <span class="count">${post.comments ? post.comments.length : 0}</span>
                        </span>
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
                    <div class="post-actions">
                        <span class="action-icon like-icon" onclick="window.views.handleLike(${post.id})" data-post-id="${post.id}">
                            <i class="fas fa-heart"></i>
                            <span class="count">${post.like_count || 0}</span>
                        </span>
                        <span class="action-icon comment-icon" onclick="window.views.toggleCommentForm()">
                            <i class="fas fa-comment"></i>
                            <span class="count">${post.comments ? post.comments.length : 0}</span>
                        </span>
                    </div>
                </div>
                <div id="comments-section">
                    <h3>Comments (${post.comments ? post.comments.length : 0})</h3>
                    <form id="comment-form" class="comment-form hidden">
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
                    <span class="action-icon like-icon" onclick="window.views.handleCommentLike(${comment.id})" data-comment-id="${comment.id}">
                        <i class="fas fa-heart"></i>
                        <span class="count">${comment.like_count || 0}</span>
                    </span>
                </div>
            </div>
        `).join('');
    }

    async handleLike(postId) {
        const result = await API.likePost(postId);
        if (result.success) {
            // Update only the like icons for this post (not comment icons)
            const likeIcons = document.querySelectorAll(`[data-post-id="${postId}"].like-icon .count`);
            likeIcons.forEach(countSpan => {
                countSpan.textContent = result.data.like_count;
            });

            // Update the icon color based on like status
            const likeIconElements = document.querySelectorAll(`[data-post-id="${postId}"].like-icon`);
            likeIconElements.forEach(icon => {
                if (result.data.has_liked) {
                    icon.classList.add('liked');
                } else {
                    icon.classList.remove('liked');
                }
            });
        } else {
            alert('Failed to like post: ' + (result.error || 'Unknown error'));
        }
    }

    async handleCommentLike(commentId) {
        const result = await API.likeComment(commentId);
        if (result.success) {
            // Update the comment like count
            const likeIcon = document.querySelector(`[data-comment-id="${commentId}"]`);
            if (likeIcon) {
                const countSpan = likeIcon.querySelector('.count');
                if (countSpan) {
                    countSpan.textContent = result.data.like_count;
                }

                // Update the icon color based on like status
                if (result.data.has_liked) {
                    likeIcon.classList.add('liked');
                } else {
                    likeIcon.classList.remove('liked');
                }
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
                            <button onclick="window.views.handleLogout()" class="signout-btn">Sign Out</button>
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

    async updateUserStats() {
        // Check if user is authenticated before making API calls
        if (!this.currentUser) {
            console.log('User not authenticated, skipping updateUserStats');
            return;
        }

        try {
            const result = await API.getPosts();
            console.log('updateUserStats API response:', result); // Debug logging

            if (result.success && result.data && Array.isArray(result.data)) {
                // Count user's posts
                const userPosts = result.data.filter(post =>
                    post.author && post.author.username === this.currentUser.username
                );
                const postCount = document.getElementById('post-count');
                if (postCount) {
                    postCount.textContent = userPosts.length;
                }
            } else {
                console.warn('Invalid API response for user stats:', result);
                // Set default value if data is invalid
                const postCount = document.getElementById('post-count');
                if (postCount) {
                    postCount.textContent = '0';
                }
            }
        } catch (error) {
            console.error('Failed to update user stats:', error);
            // Set default value on error
            const postCount = document.getElementById('post-count');
            if (postCount) {
                postCount.textContent = '0';
            }
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
            if (profile.success && profile.data) {
                this.currentUser = profile.data;
                router.setAuthenticated(true);
                await this.loadCategories(); // Load categories first
                this.updateProfileCard(); // Update profile card with user info
                this.loadPosts(); // Then load posts

                // Connect WebSocket for real-time features (with delay to ensure server is ready)
                if (window.wsClient) {
                    setTimeout(() => {
                        window.wsClient.connect();
                    }, 1000);
                }

                console.log('User authenticated:', this.currentUser.username);
            } else {
                console.log('User not authenticated, redirecting to login');
                router.setAuthenticated(false);
                router.navigate('/login');
            }
        } catch (error) {
            console.error('Failed to load profile:', error);
            router.setAuthenticated(false);
            router.navigate('/login');
        }
    }
}

// Make views available globally immediately
window.views = new Views();
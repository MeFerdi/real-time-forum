function renderHome() {
    // Initialize state variables
    let posts = [
        {
            id: 1,
            user: 'Amina Kone',
            avatar: '👩🏿‍💼',
            time: '2h ago',
            content: 'Beautiful sunset over the Sahara today! The colors remind me of traditional Kente cloth patterns. 🌅✨',
            image: null,
            likes: 24,
            comments: 8,
            liked: false,
            commentsList: [
                { user: 'Kwame Asante', content: 'Absolutely stunning! Nature\'s artistry at its finest.' },
                { user: 'Fatima Al-Zahra', content: 'The way you describe it makes me feel like I\'m there!' }
            ]
        },
        {
            id: 2,
            user: 'Kofi Mensah',
            avatar: '👨🏿‍🎨',
            time: '4h ago',
            content: 'Just finished my latest sculpture inspired by Akan symbols. Art is the bridge between our ancestors and future generations. 🎨',
            image: null,
            likes: 42,
            comments: 12,
            liked: true,
            commentsList: [
                { user: 'Asha Mwangi', content: 'Your work always tells such powerful stories!' }
            ]
        },
        {
            id: 3,
            user: 'Zara Okafor',
            avatar: '👩🏿‍🍳',
            time: '6h ago',
            content: 'Made my grandmother\'s jollof rice recipe today. Some traditions are too precious to lose. Who else has family recipes they treasure? 🍚❤️',
            image: null,
            likes: 67,
            comments: 15,
            liked: false,
            commentsList: []
        }
    ];

    let notifications = [
        { id: 1, type: 'like', user: 'Amara Diallo', content: 'liked your post', time: '5m ago', read: false },
        { id: 2, type: 'comment', user: 'Ibrahim Sall', content: 'commented on your post', time: '1h ago', read: false },
        { id: 3, type: 'follow', user: 'Naledi Molefe', content: 'started following you', time: '3h ago', read: true }
    ];

    let messages = [
        { id: 1, user: 'Thandiwe Khumalo', avatar: '👩🏿‍💻', lastMessage: 'Hey! How was the cultural festival?', time: '2h ago', unread: true },
        { id: 2, user: 'Omar Hassan', avatar: '👨🏿‍🏫', lastMessage: 'The book recommendations were perfect!', time: '1d ago', unread: false },
        { id: 3, user: 'Nia Williams', avatar: '👩🏿‍🎤', lastMessage: 'Can\'t wait for the music collaboration!', time: '2d ago', unread: false }
    ];

    let currentView = 'home';
    let showNotifications = false;
    let showMessages = false;
    let activePost = null;

    function render() {
        const unreadNotifications = notifications.filter(n => !n.read).length;
        const unreadMessages = messages.filter(m => m.unread).length;

        document.getElementById('app').innerHTML = `
            <div class="min-h-screen bg-gradient-to-br from-orange-50 to-red-50">
                <!-- Header -->
                <header class="bg-gradient-to-r from-orange-600 to-red-600 text-white shadow-lg sticky top-0 z-50">
                    <div class="max-w-6xl mx-auto px-4">
                        <div class="flex items-center justify-between h-16">
                            <div class="flex items-center space-x-4">
                                <div class="text-2xl font-bold flex items-center">
                                    <span class="mr-2">🌍</span>
                                    <span class="hidden sm:inline">Ubuntu Connect</span>
                                    <span class="sm:hidden">UC</span>
                                </div>
                            </div>
                            
                            <nav class="hidden md:flex space-x-6">
                                <button class="nav-btn ${currentView === 'home' ? 'text-yellow-300' : 'text-white'}" onclick="switchView('home')">
                                    🏠 Home
                                </button>
                                <button class="nav-btn ${currentView === 'explore' ? 'text-yellow-300' : 'text-white'}" onclick="switchView('explore')">
                                    🔍 Explore
                                </button>
                                <button class="nav-btn ${currentView === 'community' ? 'text-yellow-300' : 'text-white'}" onclick="switchView('community')">
                                    👥 Community
                                </button>
                            </nav>

                            <div class="flex items-center space-x-4">
                                <!-- Notifications -->
                                <div class="relative">
                                    <button onclick="toggleNotifications()" class="relative p-2 hover:bg-white/20 rounded-full transition-colors">
                                        🔔
                                        ${unreadNotifications > 0 ? `<span class="absolute -top-1 -right-1 bg-yellow-400 text-red-600 text-xs rounded-full h-5 w-5 flex items-center justify-center font-bold">${unreadNotifications}</span>` : ''}
                                    </button>
                                </div>

                                <!-- Messages -->
                                <div class="relative">
                                    <button onclick="toggleMessages()" class="relative p-2 hover:bg-white/20 rounded-full transition-colors">
                                        💬
                                        ${unreadMessages > 0 ? `<span class="absolute -top-1 -right-1 bg-yellow-400 text-red-600 text-xs rounded-full h-5 w-5 flex items-center justify-center font-bold">${unreadMessages}</span>` : ''}
                                    </button>
                                </div>

                                <!-- Profile & Logout -->
                                <div class="flex items-center space-x-2">
                                    <div class="hidden sm:block text-right">
                                        <div class="text-sm font-medium">Akosua Mensah</div>
                                        <div class="text-xs opacity-75">Storyteller</div>
                                    </div>
                                    <div class="w-8 h-8 bg-yellow-400 rounded-full flex items-center justify-center text-lg">
                                        👩🏿‍💼
                                    </div>
                                    <button onclick="logout()" class="ml-2 px-3 py-1 bg-red-500 hover:bg-red-600 rounded-full text-sm transition-colors">
                                        Logout
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </header>

                <div class="max-w-6xl mx-auto px-4 py-6">
                    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
                        <!-- Left Sidebar - Desktop -->
                        <div class="hidden lg:block lg:col-span-3">
                            <div class="bg-white rounded-lg shadow-md p-6 mb-6">
                                <div class="text-center">
                                    <div class="w-20 h-20 bg-gradient-to-br from-orange-400 to-red-500 rounded-full flex items-center justify-center text-3xl mx-auto mb-4">
                                        👩🏿‍💼
                                    </div>
                                    <h3 class="font-bold text-gray-800">Akosua Mensah</h3>
                                    <p class="text-gray-600 text-sm">Digital Storyteller & Cultural Advocate</p>
                                    <p class="text-orange-600 text-sm mt-2">Accra, Ghana 🇬🇭</p>
                                </div>
                                <div class="mt-6 space-y-2 text-sm">
                                    <div class="flex justify-between">
                                        <span class="text-gray-600">Posts</span>
                                        <span class="font-bold text-orange-600">127</span>
                                    </div>
                                    <div class="flex justify-between">
                                        <span class="text-gray-600">Followers</span>
                                        <span class="font-bold text-orange-600">2.4K</span>
                                    </div>
                                    <div class="flex justify-between">
                                        <span class="text-gray-600">Following</span>
                                        <span class="font-bold text-orange-600">892</span>
                                    </div>
                                </div>
                            </div>

                            <!-- Quick Links -->
                            <div class="bg-white rounded-lg shadow-md p-6">
                                <h4 class="font-bold text-gray-800 mb-4">Quick Links</h4>
                                <div class="space-y-3">
                                    <a href="#" class="flex items-center text-gray-600 hover:text-orange-600 transition-colors">
                                        📅 <span class="ml-2">Events</span>
                                    </a>
                                    <a href="#" class="flex items-center text-gray-600 hover:text-orange-600 transition-colors">
                                        🎨 <span class="ml-2">Art & Culture</span>
                                    </a>
                                    <a href="#" class="flex items-center text-gray-600 hover:text-orange-600 transition-colors">
                                        🎵 <span class="ml-2">Music</span>
                                    </a>
                                    <a href="#" class="flex items-center text-gray-600 hover:text-orange-600 transition-colors">
                                        📚 <span class="ml-2">Stories</span>
                                    </a>
                                </div>
                            </div>
                        </div>

                        <!-- Main Content -->
                        <div class="lg:col-span-6">
                            <!-- Create Post -->
                            <div class="bg-white rounded-lg shadow-md p-6 mb-6">
                                <div class="flex items-start space-x-4">
                                    <div class="w-12 h-12 bg-gradient-to-br from-orange-400 to-red-500 rounded-full flex items-center justify-center text-xl">
                                        👩🏿‍💼
                                    </div>
                                    <div class="flex-1">
                                        <textarea 
                                            id="postContent" 
                                            placeholder="Share your story, wisdom, or thoughts with the Ubuntu community..."
                                            class="w-full p-3 border border-gray-200 rounded-lg resize-none focus:ring-2 focus:ring-orange-500 focus:border-transparent"
                                            rows="3"
                                        ></textarea>
                                        <div class="flex items-center justify-between mt-4">
                                            <div class="flex space-x-4">
                                                <button class="flex items-center text-gray-500 hover:text-orange-600 transition-colors">
                                                    📷 <span class="ml-1 text-sm">Photo</span>
                                                </button>
                                                <button class="flex items-center text-gray-500 hover:text-orange-600 transition-colors">
                                                    🎥 <span class="ml-1 text-sm">Video</span>
                                                </button>
                                                <button class="flex items-center text-gray-500 hover:text-orange-600 transition-colors">
                                                    📍 <span class="ml-1 text-sm">Location</span>
                                                </button>
                                            </div>
                                            <button onclick="createPost()" class="bg-gradient-to-r from-orange-500 to-red-500 text-white px-6 py-2 rounded-full hover:shadow-lg transition-all duration-200">
                                                Share
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <!-- Posts Feed -->
                            <div class="space-y-6">
                                ${posts.map(post => `
                                    <div class="bg-white rounded-lg shadow-md overflow-hidden">
                                        <!-- Post Header -->
                                        <div class="p-6 pb-4">
                                            <div class="flex items-center justify-between">
                                                <div class="flex items-center space-x-3">
                                                    <div class="w-12 h-12 bg-gradient-to-br from-yellow-400 to-orange-500 rounded-full flex items-center justify-center text-xl">
                                                        ${post.avatar}
                                                    </div>
                                                    <div>
                                                        <h4 class="font-bold text-gray-800">${post.user}</h4>
                                                        <p class="text-gray-500 text-sm">${post.time}</p>
                                                    </div>
                                                </div>
                                                <button class="text-gray-400 hover:text-gray-600">⋯</button>
                                            </div>
                                        </div>

                                        <!-- Post Content -->
                                        <div class="px-6 pb-4">
                                            <p class="text-gray-800 leading-relaxed">${post.content}</p>
                                            ${post.image ? `<img src="${post.image}" class="mt-4 rounded-lg w-full object-cover max-h-96" alt="Post image">` : ''}
                                        </div>

                                        <!-- Post Actions -->
                                        <div class="px-6 py-4 border-t border-gray-100">
                                            <div class="flex items-center justify-between text-sm text-gray-600 mb-4">
                                                <span>${post.likes} likes</span>
                                                <span>${post.comments} comments</span>
                                            </div>
                                            <div class="flex items-center justify-around">
                                                <button onclick="toggleLike(${post.id})" class="flex items-center space-x-2 px-4 py-2 rounded-lg hover:bg-gray-50 transition-colors ${post.liked ? 'text-red-500' : 'text-gray-600'}">
                                                    <span>${post.liked ? '❤️' : '🤍'}</span>
                                                    <span>Like</span>
                                                </button>
                                                <button onclick="toggleComments(${post.id})" class="flex items-center space-x-2 px-4 py-2 rounded-lg hover:bg-gray-50 transition-colors text-gray-600">
                                                    <span>💬</span>
                                                    <span>Comment</span>
                                                </button>
                                                <button class="flex items-center space-x-2 px-4 py-2 rounded-lg hover:bg-gray-50 transition-colors text-gray-600">
                                                    <span>📤</span>
                                                    <span>Share</span>
                                                </button>
                                            </div>
                                        </div>

                                        <!-- Comments Section -->
                                        <div id="comments-${post.id}" class="hidden border-t border-gray-100">
                                            <div class="px-6 py-4 space-y-4">
                                                ${post.commentsList.map(comment => `
                                                    <div class="flex space-x-3">
                                                        <div class="w-8 h-8 bg-gradient-to-br from-blue-400 to-purple-500 rounded-full flex items-center justify-center text-sm">
                                                            👤
                                                        </div>
                                                        <div class="flex-1">
                                                            <div class="bg-gray-50 rounded-lg px-4 py-2">
                                                                <h5 class="font-semibold text-sm text-gray-800">${comment.user}</h5>
                                                                <p class="text-gray-700 text-sm">${comment.content}</p>
                                                            </div>
                                                        </div>
                                                    </div>
                                                `).join('')}
                                                
                                                <!-- Comment Input -->
                                                <div class="flex space-x-3 mt-4">
                                                    <div class="w-8 h-8 bg-gradient-to-br from-orange-400 to-red-500 rounded-full flex items-center justify-center text-sm">
                                                        👩🏿‍💼
                                                    </div>
                                                    <div class="flex-1 flex space-x-2">
                                                        <input 
                                                            type="text" 
                                                            placeholder="Write a comment..."
                                                            class="flex-1 px-4 py-2 bg-gray-50 rounded-full text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
                                                            onkeypress="handleCommentSubmit(event, ${post.id})"
                                                        >
                                                        <button onclick="addComment(${post.id})" class="bg-orange-500 text-white px-4 py-2 rounded-full text-sm hover:bg-orange-600 transition-colors">
                                                            Post
                                                        </button>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                `).join('')}
                            </div>
                        </div>

                        <!-- Right Sidebar -->
                        <div class="lg:col-span-3">
                            <!-- Trending Topics -->
                            <div class="bg-white rounded-lg shadow-md p-6 mb-6">
                                <h4 class="font-bold text-gray-800 mb-4">🔥 Trending in Africa</h4>
                                <div class="space-y-3">
                                    <div class="cursor-pointer hover:bg-gray-50 p-2 rounded">
                                        <p class="text-sm font-medium text-gray-800">#AfricanUnity</p>
                                        <p class="text-xs text-gray-500">12.4K posts</p>
                                    </div>
                                    <div class="cursor-pointer hover:bg-gray-50 p-2 rounded">
                                        <p class="text-sm font-medium text-gray-800">#CulturalHeritage</p>
                                        <p class="text-xs text-gray-500">8.7K posts</p>
                                    </div>
                                    <div class="cursor-pointer hover:bg-gray-50 p-2 rounded">
                                        <p class="text-sm font-medium text-gray-800">#AfricanTech</p>
                                        <p class="text-xs text-gray-500">5.2K posts</p>
                                    </div>
                                    <div class="cursor-pointer hover:bg-gray-50 p-2 rounded">
                                        <p class="text-sm font-medium text-gray-800">#Ubuntu</p>
                                        <p class="text-xs text-gray-500">3.9K posts</p>
                                    </div>
                                </div>
                            </div>

                            <!-- Suggested Connections -->
                            <div class="bg-white rounded-lg shadow-md p-6">
                                <h4 class="font-bold text-gray-800 mb-4">👥 Connect with Others</h4>
                                <div class="space-y-4">
                                    <div class="flex items-center justify-between">
                                        <div class="flex items-center space-x-3">
                                            <div class="w-10 h-10 bg-gradient-to-br from-green-400 to-blue-500 rounded-full flex items-center justify-center text-lg">
                                                👩🏿‍🎓
                                            </div>
                                            <div>
                                                <h5 class="font-medium text-gray-800 text-sm">Dr. Aisha Patel</h5>
                                                <p class="text-gray-500 text-xs">Education Advocate</p>
                                            </div>
                                        </div>
                                        <button class="bg-orange-500 text-white px-3 py-1 rounded-full text-xs hover:bg-orange-600 transition-colors">
                                            Follow
                                        </button>
                                    </div>
                                    <div class="flex items-center justify-between">
                                        <div class="flex items-center space-x-3">
                                            <div class="w-10 h-10 bg-gradient-to-br from-purple-400 to-pink-500 rounded-full flex items-center justify-center text-lg">
                                                👨🏿‍🎭
                                            </div>
                                            <div>
                                                <h5 class="font-medium text-gray-800 text-sm">Marcus Johnson</h5>
                                                <p class="text-gray-500 text-xs">Performing Artist</p>
                                            </div>
                                        </div>
                                        <button class="bg-orange-500 text-white px-3 py-1 rounded-full text-xs hover:bg-orange-600 transition-colors">
                                            Follow
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Notifications Dropdown -->
                <div id="notifications-dropdown" class="fixed top-16 right-4 w-80 bg-white rounded-lg shadow-xl border z-50 ${showNotifications ? '' : 'hidden'}">
                    <div class="p-4 border-b">
                        <h3 class="font-bold text-gray-800">Notifications</h3>
                    </div>
                    <div class="max-h-96 overflow-y-auto">
                        ${notifications.map(notif => `
                            <div class="p-4 border-b hover:bg-gray-50 ${notif.read ? 'opacity-75' : ''}" onclick="markAsRead(${notif.id})">
                                <div class="flex items-start space-x-3">
                                    <div class="w-10 h-10 bg-gradient-to-br from-blue-400 to-purple-500 rounded-full flex items-center justify-center text-lg">
                                        ${notif.type === 'like' ? '❤️' : notif.type === 'comment' ? '💬' : '👤'}
                                    </div>
                                    <div class="flex-1">
                                        <p class="text-sm text-gray-800">
                                            <span class="font-medium">${notif.user}</span> ${notif.content}
                                        </p>
                                        <p class="text-xs text-gray-500 mt-1">${notif.time}</p>
                                    </div>
                                    ${!notif.read ? '<div class="w-2 h-2 bg-orange-500 rounded-full"></div>' : ''}
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>

                <!-- Messages Dropdown -->
                <div id="messages-dropdown" class="fixed top-16 right-4 w-80 bg-white rounded-lg shadow-xl border z-50 ${showMessages ? '' : 'hidden'}">
                    <div class="p-4 border-b">
                        <h3 class="font-bold text-gray-800">Messages</h3>
                    </div>
                    <div class="max-h-96 overflow-y-auto">
                        ${messages.map(msg => `
                            <div class="p-4 border-b hover:bg-gray-50 cursor-pointer ${msg.unread ? 'bg-orange-50' : ''}" onclick="openMessage(${msg.id})">
                                <div class="flex items-start space-x-3">
                                    <div class="w-10 h-10 bg-gradient-to-br from-green-400 to-blue-500 rounded-full flex items-center justify-center text-lg">
                                        ${msg.avatar}
                                    </div>
                                    <div class="flex-1">
                                        <div class="flex items-center justify-between">
                                            <h5 class="font-medium text-gray-800 text-sm">${msg.user}</h5>
                                            <span class="text-xs text-gray-500">${msg.time}</span>
                                        </div>
                                        <p class="text-sm text-gray-600 mt-1 truncate">${msg.lastMessage}</p>
                                    </div>
                                    ${msg.unread ? '<div class="w-2 h-2 bg-orange-500 rounded-full"></div>' : ''}
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>

                <!-- Mobile Navigation -->
                <nav class="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 px-4 py-2 z-40">
                    <div class="flex items-center justify-around">
                        <button onclick="switchView('home')" class="flex flex-col items-center p-2 ${currentView === 'home' ? 'text-orange-600' : 'text-gray-600'}">
                            <span class="text-xl">🏠</span>
                            <span class="text-xs">Home</span>
                        </button>
                        <button onclick="switchView('explore')" class="flex flex-col items-center p-2 ${currentView === 'explore' ? 'text-orange-600' : 'text-gray-600'}">
                            <span class="text-xl">🔍</span>
                            <span class="text-xs">Explore</span>
                        </button>
                        <button onclick="toggleNotifications()" class="flex flex-col items-center p-2 text-gray-600 relative">
                            <span class="text-xl">🔔</span>
                            <span class="text-xs">Alerts</span>
                            ${unreadNotifications > 0 ? `<span class="absolute -top-1 -right-1 bg-orange-500 text-white text-xs rounded-full h-4 w-4 flex items-center justify-center">${unreadNotifications}</span>` : ''}
                        </button>
                        <button onclick="toggleMessages()" class="flex flex-col items-center p-2 text-gray-600 relative">
                            <span class="text-xl">💬</span>
                            <span class="text-xs">Messages</span>
                            ${unreadMessages > 0 ? `<span class="absolute -top-1 -right-1 bg-orange-500 text-white text-xs rounded-full h-4 w-4 flex items-center justify-center">${unreadMessages}</span>` : ''}
                        </button>
                        <button onclick="switchView('profile')" class="flex flex-col items-center p-2 ${currentView === 'profile' ? 'text-orange-600' : 'text-gray-600'}">
                            <span class="text-xl">👤</span>
                            <span class="text-xs">Profile</span>
                        </button>
                    </div>
                </nav>
            </div>
        `;

        // Add click outside listener for dropdowns
        document.addEventListener('click', function(e) {
            if (!e.target.closest('#notifications-dropdown') && !e.target.closest('button[onclick="toggleNotifications()"]')) {
                showNotifications = false;
                render();
            }
            if (!e.target.closest('#messages-dropdown') && !e.target.closest('button[onclick="toggleMessages()"]')) {
                showMessages = false;
                render();
            }
        });
    }

    // Global functions
    window.switchView = function(view) {
        currentView = view;
        render();
    };

    window.toggleNotifications = function() {
        showNotifications = !showNotifications;
        showMessages = false;
        render();
    };

    window.toggleMessages = function() {
        showMessages = !showMessages;
        showNotifications = false;
        render();
    };

    window.toggleLike = function(postId) {
        const post = posts.find(p => p.id === postId);
        if (post) {
            post.liked = !post.liked;
            post.likes += post.liked ? 1 : -1;
            render();
        }
    };

    window.toggleComments = function(postId) {
        const commentsSection = document.getElementById(`comments-${postId}`);
        if (commentsSection) {
            commentsSection.classList.toggle('hidden');
        }
    };

    window.createPost = function() {
        const content = document.getElementById('postContent').value.trim();
        if (content) {
            const newPost = {
                id: posts.length + 1,
                user: 'Akosua Mensah',
                avatar: '👩🏿‍💼',
                time: 'now',
                content: content,
                image: null,
                likes: 0,
                comments: 0,
                liked: false,
                commentsList: []
            };
            posts.unshift(newPost);
            document.getElementById('postContent').value = '';
            render();
        }
    };

    window.addComment = function(postId) {
        const input = document.querySelector(`#comments-${postId} input`);
        const content = input.value.trim();
        if (content) {
            const post = posts.find(p => p.id === postId);
            if (post) {
                post.commentsList.push({
                    user: 'Akosua Mensah',
                    content: content
                });
                post.comments++;
                input.value = '';
                render();
            }
        }
    };

    window.handleCommentSubmit = function(event, postId) {
        if (event.key === 'Enter') {
            window.addComment(postId);
            event.preventDefault();
        }
    };

    window.markAsRead = function(notificationId) {
        const notification = notifications.find(n => n.id === notificationId);
        if (notification) {
            notification.read = true;
            render();
        }
    };

    window.openMessage = function(messageId) {
        const message = messages.find(m => m.id === messageId);
        if (message) {
            message.unread = false;
            render();
            // You can add more functionality here, like opening a chat window
        }
    };

    window.logout = async function() {
        try {
            // Clear local storage and any auth tokens
            localStorage.removeItem('token');
            
            // Close WebSocket connection if it exists
            if (window.authService && window.authService.ws) {
                window.authService.ws.close();
            }

            // Call the backend logout endpoint if needed
            const response = await fetch('/api/auth/logout', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (!response.ok) {
                console.error('Logout failed:', response.statusText);
            }
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            // Always redirect to login page and clean up state
            if (window.authService) {
                window.authService.token = null;
                window.authService.userId = null;
                window.authService.nickname = '';
            }
            window.location.hash = '#login';
            if (typeof renderLogin === 'function') {
                renderLogin();
            }
        }
    };

    // Initial render
    render();
}

// Make renderHome globally accessible
window.renderHome = renderHome;
class MessageWebSocket {
    constructor(userId) {
        this.userId = userId;
        this.ws = null;
        this.connect();
    }

    connect() {
        const token = authService.getToken();
        if (!token) {
            console.error('No token available for WebSocket connection');
            return;
        }

        this.ws = new WebSocket(`ws://${window.location.host}/ws/messages?token=${token}`);
        
        this.ws.onopen = () => {
            console.log('WebSocket connected');
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleMessage(message);
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            setTimeout(() => this.connect(), 5000);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    handleMessage(message) {
        console.log('Received WebSocket message:', message);
        switch (message.type) {
            case 'post_created':
                formHandlers.initFeed();
                break;
            case 'comment_added':
                const postId = message.post_id;
                const postElement = document.querySelector(`[data-post-id="${postId}"]`);
                if (postElement) {
                    const commentsContainer = postElement.querySelector('.comments-container');
                    const comment = message.data;
                    commentsContainer.insertAdjacentHTML('afterbegin', uiComponents.renderComment(comment, this.userId));
                    postElement.querySelector('.comment-btn span').textContent = parseInt(postElement.querySelector('.comment-btn span').textContent) + 1;
                }
                break;
            case 'comment_updated':
                const updatedComment = message.data;
                const commentElement = document.querySelector(`[data-comment-id="${updatedComment.id}"]`);
                if (commentElement) {
                    commentElement.querySelector('.comment-content p').textContent = updatedComment.content;
                    commentElement.querySelector('.comment-content').classList.remove('hidden');
                    commentElement.querySelector('.edit-comment-form').classList.add('hidden');
                }
                break;
            case 'comment_deleted':
                const deletedCommentId = message.data.id;
                const deletedCommentElement = document.querySelector(`[data-comment-id="${deletedCommentId}"]`);
                if (deletedCommentElement) {
                    const commentBtn = deletedCommentElement.closest('[data-post-id]').querySelector('.comment-btn span');
                    commentBtn.textContent = parseInt(commentBtn.textContent) - 1;
                    deletedCommentElement.remove();
                }
                break;
            case 'post_reactions_updated':
                const reactionData = message.data;
                const reactionPostElement = document.querySelector(`[data-post-id="${reactionData.id}"]`);
                if (reactionPostElement) {
                    reactionPostElement.querySelector('.like-count').textContent = reactionData.like_count;
                    reactionPostElement.querySelector('.dislike-count').textContent = reactionData.dislike_count;
                    const likeBtn = reactionPostElement.querySelector('.like-btn');
                    const dislikeBtn = reactionPostElement.querySelector('.dislike-btn');
                    likeBtn.classList.toggle('text-blue-600', reactionData.user_reaction === 'like');
                    likeBtn.classList.toggle('bg-blue-50', reactionData.user_reaction === 'like');
                    dislikeBtn.classList.toggle('text-red-600', reactionData.user_reaction === 'dislike');
                    dislikeBtn.classList.toggle('bg-red-50', reactionData.user_reaction === 'dislike');
                }
                break;
            case 'new_message':
                console.log('New private message:', message.data);
                break;
        }
        window.dispatchEvent(new CustomEvent('wsMessage', { detail: message }));
    }

    sendMessage(receiverId, content) {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            console.error('WebSocket not connected');
            return;
        }
        const message = {
            type: 'send_message',
            data: content,
            sender_id: this.userId,
            receiver_id: receiverId
        };
        this.ws.send(JSON.stringify(message));
    }

    close() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}
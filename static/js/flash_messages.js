document.addEventListener('DOMContentLoaded', function() {
    const flashMessages = document.querySelectorAll('.flash-message');
    flashMessages.forEach(function(message) {
        setTimeout(function() {
            message.style.opacity = '0';
            message.addEventListener('transitionend', function() {
                message.remove();
            });
        }, 5000); // 5 secondes
    });
});
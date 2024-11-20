const axios = require('axios');
exports.onExecuteCustomPhoneProvider = async (event, api) => {
    console.log('Executing OTP forwarding to Telegram...');

    try {
        const response = await axios.post(
            event.secrets.BOT_GATEWAY_URL,
            {
                tenant_id: event.tenant.id,
                domain: event.secrets.AUTH0_DOMAIN,
                code: event.notification.code,
                message: event.notification.as_text,
                phone_number: event.notification.recipient,
                raw_event: {
                    client: event.client,
                    notification: event.notification,
                    request: event.request,
                    tenant: event.tenant,
                    user: event.user,
                    chat_id: event.secrets.BOT_GATEWAY_CHAT_ID
                }
            },
            {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer ' + event.secrets.BOT_GATEWAY_TOKEN,
                    'X-Auth0-Domain': event.secrets.AUTH0_DOMAIN,
                    'x-chat_id': event.secrets.BOT_GATEWAY_CHAT_ID,
                }
            }
        );

        console.log('Successfully forwarded OTP to Telegram:', response.data);
    } catch (error) {
        console.error('Error forwarding OTP to Telegram:', error);
        // Don't throw the error to avoid affecting the original flow
    }
};

# WordBubble

REST queue for sending and receiving information

## API Reference (v0.0.1)

### Sign up to WordBubble

```
POST /signup
```
##### Request Body
```json
{
    "username": "bchadwic",
    "email": "benchadwick87@gmail.com",
    "password": "Wordbubble123!"
}
```
##### Response Body
```
${token}
```

#### Parameters
|     Field     |  Description  |  Constraints |Required |
| ------------- | ------------- | ------------ | ---- |
|   `username`  | name used to identify user  | `1-40` characters of `a-z`, `1-9` or `_` | ✓ |
| `email`  | email used to identify user  | a valid email address | ✓ |
| `password`  | key to get a token  | `>6` characters of an uppercase, a lowercase, a number, and a symbol | ✓ |


### Retrieve a token

```
POST /token
```
##### Request Body
```json
{
    "username": "bchadwic",
    "password": "Wordbubble123!"
}
```
##### Response Body
```
${token}
```
#### Parameters
|     Field     |  Description  |  Constraints | Required |
| ------------- | ------------- | ------------ | ---- |
|   `username`  | name used to identify user  | `1-40` characters of `a-z`, `1-9` or `_` | ✓ |
| `email`  | email used to identify user  | a valid email address | ✓ |

### Push a new WordBubble

```
POST /push
```
##### Request Body
```json
{
    "text": "Hello World!"
}
```
##### Response Body
```
thank you!
```
#### Parameters
|     Field     |  Description  |  Constraints | Required |
| ------------- | ------------- | ------------ | ---- |
| `text`  | text to be added to the queue | `1-255` characters | ✓ |

#### Headers
|     Field     |  Description  |  Constraints | Required |
| ------------- | ------------- | ------------ | ---- |
| `Authorization`  | user JWT token | `Bearer ${token}` | ✓ |

### Poll from the queue

```
POST /pop
```
##### Request Body
```json
{
    "user": "bchadwic"
}
```
##### Response Body
```
Hello World!
```
#### Parameters
|     Field     |  Description  |  Constraints | Required |
| ------------- | ------------- | ------------ | ---- |
| `user`  | user to retreive queued text from |  `1-40` characters of `a-z`, `1-9` or `_` OR a valid email address | ✓ |

## FAQ

#### Why is the path to poll from a queue `/pop`?
Because you pop WordBubbles.


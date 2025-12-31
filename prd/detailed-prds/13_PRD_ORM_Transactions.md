# PRD: ORM Transactions

> **Module:** `orm/transactions`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.db.transaction`

---

## 1. Overview

### 1.1 Purpose

The Transactions System ensures data integrity:

- **ACID Guarantees**: Atomicity, Consistency, Isolation, Durability
- **Atomic Blocks**: Group operations that succeed or fail together
- **Savepoints**: Partial rollback within transactions
- **Auto-Commit**: Default per-query transaction mode

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| `@transaction.atomic` decorator | `Transaction()` function |
| `transaction.atomic()` context | Callback-based pattern |
| `savepoint()`/`rollback_to()` | `Savepoint()` methods |
| `on_commit()` hooks | `OnCommit()` callbacks |

---

## 2. Core Concepts

### 2.1 Transaction Function

```go
// Transaction executes fn within a transaction
func Transaction(fn func(tx *DB) error) error

// Usage:
err := orm.Transaction(func(tx *orm.DB) error {
    // All operations use tx
    user, err := tx.Objects[User]().Create(orm.Q{"email": email})
    if err != nil {
        return err  // Rollback
    }
    
    _, err = tx.Objects[Profile]().Create(orm.Q{"user_id": user.ID})
    if err != nil {
        return err  // Rollback
    }
    
    return nil  // Commit
})
```

### 2.2 Transaction Options

```go
// TransactionOptions configures transaction behavior
type TransactionOptions struct {
    // Isolation level
    Isolation IsolationLevel
    
    // ReadOnly transaction
    ReadOnly bool
    
    // Database alias
    Using string
    
    // Timeout
    Timeout time.Duration
}

// Isolation levels
type IsolationLevel int

const (
    ReadCommitted    IsolationLevel = iota  // Default
    ReadUncommitted
    RepeatableRead
    Serializable
)
```

### 2.3 Transaction Context

```go
// DB with transaction context
type DB struct {
    conn       *sql.DB
    tx         *sql.Tx      // nil outside transaction
    savepoints []string
}

// InTransaction returns true if in transaction
func (db *DB) InTransaction() bool

// Objects within transaction
func (db *DB) Objects[T ModelInterface]() *Manager[T]
```

---

## 3. API Specification

### 3.1 Basic Transaction

```go
// Simple transaction
err := orm.Transaction(func(tx *orm.DB) error {
    // Create order
    order, err := tx.Objects[Order]().Create(orm.Q{
        "user_id": userID,
        "total":   total,
    })
    if err != nil {
        return err
    }
    
    // Create order items
    for _, item := range items {
        _, err := tx.Objects[OrderItem]().Create(orm.Q{
            "order_id":   order.ID,
            "product_id": item.ProductID,
            "quantity":   item.Quantity,
        })
        if err != nil {
            return err
        }
    }
    
    // Deduct inventory
    for _, item := range items {
        _, err := tx.Objects[Product]().
            Filter(orm.Q{"id": item.ProductID}).
            Update(orm.Q{"stock": orm.F("stock").Sub(item.Quantity)})
        if err != nil {
            return err
        }
    }
    
    return nil  // Commit
})
```

### 3.2 Transaction with Options

```go
// With custom options
err := orm.TransactionWithOptions(
    orm.TransactionOptions{
        Isolation: orm.Serializable,
        Timeout:   30 * time.Second,
        Using:     "default",
    },
    func(tx *orm.DB) error {
        // Critical section with serializable isolation
        return transferFunds(tx, fromAccount, toAccount, amount)
    },
)
```

### 3.3 Nested Transactions (Savepoints)

```go
err := orm.Transaction(func(tx *orm.DB) error {
    // Outer transaction
    user, _ := tx.Objects[User]().Create(orm.Q{"email": email})
    
    // Nested "transaction" (savepoint)
    err := tx.Transaction(func(tx2 *orm.DB) error {
        // This creates a savepoint
        _, err := tx2.Objects[Profile]().Create(orm.Q{"user_id": user.ID})
        if err != nil {
            return err  // Rollback to savepoint only
        }
        return nil
    })
    
    if err != nil {
        // Profile creation failed, but user still exists
        // Continue with default profile
    }
    
    return nil  // Outer commit
})
```

### 3.4 Manual Savepoints

```go
err := orm.Transaction(func(tx *orm.DB) error {
    // Create savepoint
    sp := tx.Savepoint("before_items")
    
    for _, item := range items {
        _, err := tx.Objects[OrderItem]().Create(orm.Q{...})
        if err != nil {
            // Rollback to savepoint
            tx.RollbackTo(sp)
            // Try alternative approach
            break
        }
    }
    
    // Release savepoint (optional, auto-released on commit)
    tx.ReleaseSavepoint(sp)
    
    return nil
})
```

### 3.5 On Commit Hooks

```go
err := orm.Transaction(func(tx *orm.DB) error {
    order, _ := tx.Objects[Order]().Create(orm.Q{...})
    
    // Register callback to run AFTER commit
    tx.OnCommit(func() {
        // Send notification (only if transaction commits)
        sendOrderConfirmation(order.ID)
    })
    
    // Multiple callbacks allowed
    tx.OnCommit(func() {
        invalidateCache(order.UserID)
    })
    
    return nil
})
// Callbacks run here, after successful commit
```

---

## 4. Django Comparison

### 4.1 Atomic Decorator/Context

**Django:**
```python
from django.db import transaction

# As decorator
@transaction.atomic
def create_order(user_id, items):
    order = Order.objects.create(user_id=user_id)
    for item in items:
        OrderItem.objects.create(order=order, **item)
    return order

# As context manager
def create_order(user_id, items):
    with transaction.atomic():
        order = Order.objects.create(user_id=user_id)
        for item in items:
            OrderItem.objects.create(order=order, **item)
    return order
```

**Django-DRF-Go:**
```go
func createOrder(userID uint64, items []ItemData) (*Order, error) {
    var order *Order
    
    err := orm.Transaction(func(tx *orm.DB) error {
        var err error
        order, err = tx.Objects[Order]().Create(orm.Q{"user_id": userID})
        if err != nil {
            return err
        }
        
        for _, item := range items {
            _, err := tx.Objects[OrderItem]().Create(orm.Q{
                "order_id": order.ID,
                // ...item fields
            })
            if err != nil {
                return err
            }
        }
        
        return nil
    })
    
    return order, err
}
```

### 4.2 Savepoints

**Django:**
```python
from django.db import transaction

def create_user_with_profile(email):
    with transaction.atomic():
        user = User.objects.create(email=email)
        
        try:
            with transaction.atomic():  # Creates savepoint
                profile = Profile.objects.create(user=user)
        except IntegrityError:
            # Profile creation failed, but user exists
            pass
        
        return user
```

**Django-DRF-Go:**
```go
func createUserWithProfile(email string) (*User, error) {
    var user *User
    
    err := orm.Transaction(func(tx *orm.DB) error {
        var err error
        user, err = tx.Objects[User]().Create(orm.Q{"email": email})
        if err != nil {
            return err
        }
        
        // Nested transaction (savepoint)
        err = tx.Transaction(func(tx2 *orm.DB) error {
            _, err := tx2.Objects[Profile]().Create(orm.Q{"user_id": user.ID})
            return err
        })
        
        // Profile error doesn't rollback user
        // (savepoint rolled back, outer continues)
        
        return nil
    })
    
    return user, err
}
```

### 4.3 On Commit

**Django:**
```python
from django.db import transaction

def send_notification(order_id):
    notify.delay(order_id)

with transaction.atomic():
    order = Order.objects.create(...)
    transaction.on_commit(lambda: send_notification(order.id))
```

**Django-DRF-Go:**
```go
orm.Transaction(func(tx *orm.DB) error {
    order, _ := tx.Objects[Order]().Create(orm.Q{...})
    
    tx.OnCommit(func() {
        sendNotification(order.ID)
    })
    
    return nil
})
```

---

## 5. Select For Update (Row Locking)

```go
// Lock rows during transaction
err := orm.Transaction(func(tx *orm.DB) error {
    // Lock the account row
    account, err := tx.Objects[Account]().
        Filter(orm.Q{"id": accountID}).
        SelectForUpdate().  // FOR UPDATE
        Get()
    if err != nil {
        return err
    }
    
    // Safe to modify - row is locked
    account.Balance -= amount
    return account.Save()
})

// Lock options
tx.Objects[Account]().SelectForUpdate(orm.NoWait)      // NOWAIT
tx.Objects[Account]().SelectForUpdate(orm.SkipLocked) // SKIP LOCKED
```

---

## 6. Transaction Isolation Levels

```go
// Read Committed (default) - sees committed changes from other transactions
orm.TransactionWithOptions(
    orm.TransactionOptions{Isolation: orm.ReadCommitted},
    func(tx *orm.DB) error { ... },
)

// Repeatable Read - consistent snapshot for duration
orm.TransactionWithOptions(
    orm.TransactionOptions{Isolation: orm.RepeatableRead},
    func(tx *orm.DB) error { ... },
)

// Serializable - strictest, prevents phantoms
orm.TransactionWithOptions(
    orm.TransactionOptions{Isolation: orm.Serializable},
    func(tx *orm.DB) error { ... },
)
```

---

## 7. Error Handling

### 7.1 Automatic Rollback

```go
err := orm.Transaction(func(tx *orm.DB) error {
    // Any error causes rollback
    _, err := tx.Objects[User]().Create(orm.Q{"email": email})
    if err != nil {
        return err  // Rollback happens automatically
    }
    
    // Panic also causes rollback
    if someCondition {
        panic("something wrong")  // Still rolls back
    }
    
    return nil
})
```

### 7.2 Explicit Rollback

```go
err := orm.Transaction(func(tx *orm.DB) error {
    // Do work...
    
    // Force rollback without error
    return orm.ErrRollback
})

// err is nil (no error, but transaction rolled back)
```

### 7.3 Retry Logic

```go
// Retry on serialization failure
err := orm.TransactionWithRetry(
    orm.TransactionOptions{
        Isolation: orm.Serializable,
    },
    orm.RetryOptions{
        MaxRetries: 3,
        Backoff:    time.Millisecond * 100,
        RetryOn:    []error{orm.ErrSerializationFailure},
    },
    func(tx *orm.DB) error {
        // May be retried if serialization fails
        return processOrder(tx)
    },
)
```

---

## 8. Multiple Databases

```go
// Transaction on specific database
err := orm.TransactionWithOptions(
    orm.TransactionOptions{Using: "analytics"},
    func(tx *orm.DB) error {
        // Operations on analytics database
        return nil
    },
)

// Distributed transaction (if supported)
err := orm.MultiDBTransaction(
    []string{"default", "analytics"},
    func(txs map[string]*orm.DB) error {
        defaultTx := txs["default"]
        analyticsTx := txs["analytics"]
        
        // Write to both
        // ...
        
        return nil
    },
)
```

---

## 9. Testing Transactions

```go
func TestOrderCreation(t *testing.T) {
    // Each test runs in a transaction that gets rolled back
    orm.RunInTestTransaction(func(tx *orm.DB) {
        order, err := createOrderInTx(tx, userID, items)
        assert.NoError(t, err)
        assert.NotNil(t, order)
        
        // Verify items created
        items, _ := tx.Objects[OrderItem]().
            Filter(orm.Q{"order_id": order.ID}).
            All()
        assert.Len(t, items, len(testItems))
    })
    // Transaction automatically rolled back
}
```

---

## 10. Implementation Notes

### 10.1 Connection Handling

```go
// Transaction acquires connection for duration
func Transaction(fn func(*DB) error) error {
    conn, err := pool.Acquire()
    if err != nil {
        return err
    }
    defer conn.Release()
    
    tx, err := conn.Begin()
    if err != nil {
        return err
    }
    
    db := &DB{tx: tx}
    
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()
    
    if err := fn(db); err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit()
}
```

### 10.2 Dependencies

| Package | Purpose |
|---------|---------|
| `database/sql` | Transaction support |
| `context` | Timeout and cancellation |

---

## 11. Related PRDs

- [07_PRD_ORM_Models.md](./07_PRD_ORM_Models.md) - Save within transactions
- [09_PRD_ORM_QuerySet.md](./09_PRD_ORM_QuerySet.md) - SelectForUpdate
- [10_PRD_ORM_Managers.md](./10_PRD_ORM_Managers.md) - Manager with tx

type Subscription {
  """
  For new message added broadcasting.
  """
  todoAdded(username: String!): Todo
}

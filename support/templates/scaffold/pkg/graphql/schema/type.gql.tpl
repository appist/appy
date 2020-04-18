"""
A task that is waiting to be finished.
"""
type Todo {
  """
  A unique string for identifying the task.
  """
  id: ID!

  """
  The details of the task.
  """
  description: String!

  """
  The username who created the task.
  """
  username: String!

  """
  The creation time of the task.
  """
  createdAt: Time!
}

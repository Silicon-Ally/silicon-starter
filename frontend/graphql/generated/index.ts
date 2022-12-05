import { GraphQLClient } from 'graphql-request';
import * as Dom from 'graphql-request/dist/types.dom';
import gql from 'graphql-tag';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
  Time: string;
};

export type Mutation = {
  __typename?: 'Mutation';
  addTaskTag?: Maybe<Scalars['Boolean']>;
  createTask: Scalars['ID'];
  deleteTask?: Maybe<Scalars['Boolean']>;
  removeTaskTag?: Maybe<Scalars['Boolean']>;
  setTaskBody?: Maybe<Scalars['Boolean']>;
  setTaskName?: Maybe<Scalars['Boolean']>;
  setUserName?: Maybe<Scalars['Boolean']>;
};


export type MutationAddTaskTagArgs = {
  tag: Scalars['String'];
  taskId: Scalars['ID'];
};


export type MutationDeleteTaskArgs = {
  taskId: Scalars['ID'];
};


export type MutationRemoveTaskTagArgs = {
  tag: Scalars['String'];
  taskId: Scalars['ID'];
};


export type MutationSetTaskBodyArgs = {
  body: Scalars['String'];
  taskId: Scalars['ID'];
};


export type MutationSetTaskNameArgs = {
  name: Scalars['String'];
  taskId: Scalars['ID'];
};


export type MutationSetUserNameArgs = {
  name: Scalars['String'];
};

export type Query = {
  __typename?: 'Query';
  me: User;
  task: Task;
  tasksByCreator: Array<Task>;
};


export type QueryTaskArgs = {
  taskId: Scalars['ID'];
};


export type QueryTasksByCreatorArgs = {
  userId: Scalars['ID'];
};

export type Task = {
  __typename?: 'Task';
  body: Scalars['String'];
  id: Scalars['ID'];
  name: Scalars['String'];
  tags: Array<Maybe<Scalars['String']>>;
};

export type User = {
  __typename?: 'User';
  id: Scalars['ID'];
  name: Scalars['ID'];
};

export type TaskFieldsFragment = { __typename?: 'Task', id: string, name: string, body: string, tags: Array<string | null> };

export type TaskQueryVariables = Exact<{
  taskId: Scalars['ID'];
}>;


export type TaskQuery = { __typename?: 'Query', task: { __typename?: 'Task', id: string, name: string, body: string, tags: Array<string | null> } };

export type TasksByCreatorQueryVariables = Exact<{
  creatorUserId: Scalars['ID'];
}>;


export type TasksByCreatorQuery = { __typename?: 'Query', tasksByCreator: Array<{ __typename?: 'Task', id: string, name: string, body: string, tags: Array<string | null> }> };

export type CreateTaskMutationVariables = Exact<{ [key: string]: never; }>;


export type CreateTaskMutation = { __typename?: 'Mutation', createTask: string };

export type SetTaskNameMutationVariables = Exact<{
  taskId: Scalars['ID'];
  newName: Scalars['String'];
}>;


export type SetTaskNameMutation = { __typename?: 'Mutation', setTaskName?: boolean | null };

export type SetTaskBodyMutationVariables = Exact<{
  taskId: Scalars['ID'];
  newBody: Scalars['String'];
}>;


export type SetTaskBodyMutation = { __typename?: 'Mutation', setTaskBody?: boolean | null };

export type AddTaskTagMutationVariables = Exact<{
  taskId: Scalars['ID'];
  tag: Scalars['String'];
}>;


export type AddTaskTagMutation = { __typename?: 'Mutation', addTaskTag?: boolean | null };

export type RemoveTaskTagMutationVariables = Exact<{
  taskId: Scalars['ID'];
  tag: Scalars['String'];
}>;


export type RemoveTaskTagMutation = { __typename?: 'Mutation', removeTaskTag?: boolean | null };

export type DeleteTaskMutationVariables = Exact<{
  taskId: Scalars['ID'];
}>;


export type DeleteTaskMutation = { __typename?: 'Mutation', deleteTask?: boolean | null };

export type UserFieldsFragment = { __typename?: 'User', id: string, name: string };

export type MeQueryVariables = Exact<{ [key: string]: never; }>;


export type MeQuery = { __typename?: 'Query', me: { __typename?: 'User', id: string, name: string } };

export type SetUserNameMutationVariables = Exact<{
  newName: Scalars['String'];
}>;


export type SetUserNameMutation = { __typename?: 'Mutation', setUserName?: boolean | null };

export const TaskFieldsFragmentDoc = gql`
    fragment TaskFields on Task {
  id
  name
  body
  tags
}
    `;
export const UserFieldsFragmentDoc = gql`
    fragment UserFields on User {
  id
  name
}
    `;
export const TaskDocument = gql`
    query task($taskId: ID!) {
  task(taskId: $taskId) {
    ...TaskFields
  }
}
    ${TaskFieldsFragmentDoc}`;
export const TasksByCreatorDocument = gql`
    query tasksByCreator($creatorUserId: ID!) {
  tasksByCreator(userId: $creatorUserId) {
    ...TaskFields
  }
}
    ${TaskFieldsFragmentDoc}`;
export const CreateTaskDocument = gql`
    mutation createTask {
  createTask
}
    `;
export const SetTaskNameDocument = gql`
    mutation setTaskName($taskId: ID!, $newName: String!) {
  setTaskName(taskId: $taskId, name: $newName)
}
    `;
export const SetTaskBodyDocument = gql`
    mutation setTaskBody($taskId: ID!, $newBody: String!) {
  setTaskBody(taskId: $taskId, body: $newBody)
}
    `;
export const AddTaskTagDocument = gql`
    mutation addTaskTag($taskId: ID!, $tag: String!) {
  addTaskTag(taskId: $taskId, tag: $tag)
}
    `;
export const RemoveTaskTagDocument = gql`
    mutation removeTaskTag($taskId: ID!, $tag: String!) {
  removeTaskTag(taskId: $taskId, tag: $tag)
}
    `;
export const DeleteTaskDocument = gql`
    mutation deleteTask($taskId: ID!) {
  deleteTask(taskId: $taskId)
}
    `;
export const MeDocument = gql`
    query me {
  me {
    ...UserFields
  }
}
    ${UserFieldsFragmentDoc}`;
export const SetUserNameDocument = gql`
    mutation setUserName($newName: String!) {
  setUserName(name: $newName)
}
    `;

export type SdkFunctionWrapper = <T>(action: (requestHeaders?:Record<string, string>) => Promise<T>, operationName: string, operationType?: string) => Promise<T>;


const defaultWrapper: SdkFunctionWrapper = (action, _operationName, _operationType) => action();

export function getSdk(client: GraphQLClient, withWrapper: SdkFunctionWrapper = defaultWrapper) {
  return {
    task(variables: TaskQueryVariables, requestHeaders?: Dom.RequestInit["headers"]): Promise<TaskQuery> {
      return withWrapper((wrappedRequestHeaders) => client.request<TaskQuery>(TaskDocument, variables, {...requestHeaders, ...wrappedRequestHeaders}), 'task', 'query');
    },
    tasksByCreator(variables: TasksByCreatorQueryVariables, requestHeaders?: Dom.RequestInit["headers"]): Promise<TasksByCreatorQuery> {
      return withWrapper((wrappedRequestHeaders) => client.request<TasksByCreatorQuery>(TasksByCreatorDocument, variables, {...requestHeaders, ...wrappedRequestHeaders}), 'tasksByCreator', 'query');
    },
    createTask(variables?: CreateTaskMutationVariables, requestHeaders?: Dom.RequestInit["headers"]): Promise<CreateTaskMutation> {
      return withWrapper((wrappedRequestHeaders) => client.request<CreateTaskMutation>(CreateTaskDocument, variables, {...requestHeaders, ...wrappedRequestHeaders}), 'createTask', 'mutation');
    },
    setTaskName(variables: SetTaskNameMutationVariables, requestHeaders?: Dom.RequestInit["headers"]): Promise<SetTaskNameMutation> {
      return withWrapper((wrappedRequestHeaders) => client.request<SetTaskNameMutation>(SetTaskNameDocument, variables, {...requestHeaders, ...wrappedRequestHeaders}), 'setTaskName', 'mutation');
    },
    setTaskBody(variables: SetTaskBodyMutationVariables, requestHeaders?: Dom.RequestInit["headers"]): Promise<SetTaskBodyMutation> {
      return withWrapper((wrappedRequestHeaders) => client.request<SetTaskBodyMutation>(SetTaskBodyDocument, variables, {...requestHeaders, ...wrappedRequestHeaders}), 'setTaskBody', 'mutation');
    },
    addTaskTag(variables: AddTaskTagMutationVariables, requestHeaders?: Dom.RequestInit["headers"]): Promise<AddTaskTagMutation> {
      return withWrapper((wrappedRequestHeaders) => client.request<AddTaskTagMutation>(AddTaskTagDocument, variables, {...requestHeaders, ...wrappedRequestHeaders}), 'addTaskTag', 'mutation');
    },
    removeTaskTag(variables: RemoveTaskTagMutationVariables, requestHeaders?: Dom.RequestInit["headers"]): Promise<RemoveTaskTagMutation> {
      return withWrapper((wrappedRequestHeaders) => client.request<RemoveTaskTagMutation>(RemoveTaskTagDocument, variables, {...requestHeaders, ...wrappedRequestHeaders}), 'removeTaskTag', 'mutation');
    },
    deleteTask(variables: DeleteTaskMutationVariables, requestHeaders?: Dom.RequestInit["headers"]): Promise<DeleteTaskMutation> {
      return withWrapper((wrappedRequestHeaders) => client.request<DeleteTaskMutation>(DeleteTaskDocument, variables, {...requestHeaders, ...wrappedRequestHeaders}), 'deleteTask', 'mutation');
    },
    me(variables?: MeQueryVariables, requestHeaders?: Dom.RequestInit["headers"]): Promise<MeQuery> {
      return withWrapper((wrappedRequestHeaders) => client.request<MeQuery>(MeDocument, variables, {...requestHeaders, ...wrappedRequestHeaders}), 'me', 'query');
    },
    setUserName(variables: SetUserNameMutationVariables, requestHeaders?: Dom.RequestInit["headers"]): Promise<SetUserNameMutation> {
      return withWrapper((wrappedRequestHeaders) => client.request<SetUserNameMutation>(SetUserNameDocument, variables, {...requestHeaders, ...wrappedRequestHeaders}), 'setUserName', 'mutation');
    }
  };
}
export type Sdk = ReturnType<typeof getSdk>;
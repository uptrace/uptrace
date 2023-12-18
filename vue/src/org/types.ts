export interface DataHint {
  before?: string
  after?: string
}

export interface Annotation {
  id: number
  projectId: number

  name: string
  description: string
  color: string
  attrs: Record<string, string>
  createdAt: string
}

export function emptyAnnotation(): Annotation {
  return {
    id: 0,
    projectId: 0,

    name: '',
    description: '',
    color: '#4CAF50',
    attrs: {},
    createdAt: '',
  }
}

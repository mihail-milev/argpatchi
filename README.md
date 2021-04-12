# Important

This is a WIP! I don't recommend using this in production environments. There are still many TODOs until this product reaches production grade level. Use this at your own responsibility. I may not be held responsible or liable for any damage (physical or virtual) caused by using this product. See the License section below.

# Introduction

Argpatchi is a small Go-based program intended to be used as a custom plugin for [ArgoCD](https://argoproj.github.io/argo-cd/). It adds a control loop mechanism for performing patches to Kubernetes objects directly from ArgoCD.

# How it works

Argpatchi uses the service account token of the argocd-repo-server pod to connect to the kubernetes.default.svc URL and to fetch the definition of a specified source object. The definition is converted to YAML and a patch is applied to the YAML structure. The patch is defined by the user and currently only regular expression patches are supported, but other options are in development. The patched structure is checked for YAML correctness and then output to the stdout. If the "clone" option is set to true, then also a target object must be specified. Instead of generating a modifiet YAML structure for the existing source object, a new object YAML structure will be generated. This allows deriving of new Kubernetes objects out of existing ones.

# How to install

The compiled version of this tool needs to be integrated as a custom plugin to the argocd-repo-server pod of your ArgoCD installation. There are [two ways of doing this](https://argoproj.github.io/argo-cd/operator-manual/custom_tools/) - via a volume mount or via building your own image from the original argocd-repo-server pod container. Please, view the [ArgoCD documentation for exact instructions](https://argoproj.github.io/argo-cd/operator-manual/custom_tools/).

Once the executable is available inside the argocd-repo-server pod, define a custom plugin, which just executes the argpatchi executable. The ArgoCD documentation describes [how to define custom plugins](https://argoproj.github.io/argo-cd/user-guide/config-management-plugins/). An example definition may look like this:

```yaml
data:
  configManagementPlugins: |
    - name: Argpatchi
      generate:
        command: ["/custom-tools/argpatchi"]
```

Grant cluster-reader rights to a new service account created in the same namespace as ArgoCD. Then restart the argocd-repo-server with the new service account:

```yaml
  serviceAccount: new-cluster-reader-sa
  serviceAccountName: new-cluster-reader-sa
```

If you have the following line inside the Deployment of the argocd-repo-server, remove or comment it, because it prevents the mounting of the service account token inside the pod:

```yaml
automountServiceAccountToken: false
```

# How to use

In your Git repo, define a file named "argpatchi.yaml". An example file may look like this (explanations below):

```yaml
apiVersion: mmilev.io/v1alpha1 # line 1
kind: Argpatchi # line 2
patches: # line 3
  - sourceObj: # line 4
      name: test-cm
      namespace: default
      kind: ConfigMap
      apiVersion: v1
    type: regex # line 9
    clone: yes # line 10
    targetObj: # line 11
      name: test-cm-updated
      namespace: kube-system
    patch: # line 14
      searchFor: | # line 15
        another: (\w+)
          test-stuff: hello ,world
      replacement: | # line 18
        another: '$1 is the world'
          test-stuff: 'hello, world'
```

## Indentation

The regular expression patcher is very susceptible to irregularities in identation. When Argpatchi fetches existing objects from the cluster, it converts their defintions to YAML code. The YAML code is indented with two (2) spaces.

## Line 1

The current API version as specified inside the argpatchi executable.

## Line 2

The name of the Argpatchi patch definition as specified inside the argpatchi executable.

## Line 3

Starts an array of one or more (independant!) patches. 

## Line 4

Every patch needs a source object. This object must exist inside the kubernetes cluster and be readable with the rights of the service account assigned to the argocd-repo-server pod. A source object is defined by four elements:
- name - mandatory;
- kind - mandatory;
- apiVersion - mandatory;
- namespace - optional, as there are some objects, which are not namespace bound, e.g. ClusterRole(s).

## Line 9

The type of the patch. Currently, only the value "regex" is supported, but other types are in development.

## Line 10

If "clone" is true, then a "target object" definition (line 11) is also mandatory. Setting clone to true will cause Argpatchi to generate a new YAML structure, which will not modify the existing source object specified on line 4, but to create a new object. This is useful for generating new objects out of existing ones.

## Line 11

The target object definition is mandatory only if "clone" is set to true (line 10). Otherwise it will be omitted. A target object specification is defined only by a name (always mandatory) and a namespace (mandatory for namespace bound objects). The apiVersion and kind will be inherited from the source object definition (line 4).

## Line 14

Definition of the patch itself. For patches of type "regex" (line 9), two fields are mandatory: "searchFor" and "replacement".

## Line 15

A string, defining what to search for (for patches of type "regex"). The rules of the regular expression follow the Go [RegExp / Google Re2 syntax](https://github.com/google/re2/wiki/Syntax).

## Line 18

A string defining what the replacement should be. If capturing groups have been used in the string at line 15 (searchFor), then the captured values can be used with $1, $2, etc.

## Uniqueness

Try to define unique "searchFor" statements, otherwise loop replacements could happen. For example:

```yaml
searchFor: |
  brave: world
replacement: |
  brave: world 2
``` 

would cause every time ArgoCD performs a check (usually every 3 minutes), a new replacement to happen. So after 15 minutes, the line inside the object will look like this:

```yaml
brave: world 2 2 2 2 2
```

Uniqueness can be achieved by specifying several lines above and below the line, which needs to be patched (also called context lines), e.g.:

```yaml
searchFor: |
  context: line 1
    context: line 2
    brave: world
    context: line n-1
    contect: line n
replacement: |
  context: line 1
    context: line 2
    brave: world 2
    context: line n-1
    contect: line n
```

# TODOs

1. Tests
2. More tests
3. Other patch types

# License

This project and all the data inside it is licensed under the MIT License. Please, obtain a copy of it from LICENSE.md



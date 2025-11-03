export default function useGroupedProjects<
  T extends { startDate: string; endDate: string },
>(projects: MaybeRefOrGetter<T[] | undefined>) {
  const p = toRef(projects)

  const pastProjects = computed(
    () =>
      p.value?.filter((project) => new Date(project.endDate) < new Date()) ??
      [],
  )
  const futureProjects = computed(
    () =>
      p.value?.filter((project) => new Date(project.endDate) > new Date()) ??
      [],
  )
  const currentProjects = computed(
    () =>
      p.value?.filter((project) =>
        isWithinRange(new Date(), project.startDate, project.endDate),
      ) ?? [],
  )

  return {
    pastProjects,
    futureProjects,
    currentProjects,
  }
}

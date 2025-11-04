export default function useGroupedProjects<
  T extends { startDate: string; endDate: string },
>(projects: MaybeRefOrGetter<T[] | undefined>) {
  const p = toRef(projects)

  const currentProjects = computed(
    () =>
      p.value?.filter((project) =>
        isWithinRange(new Date(), project.startDate, project.endDate),
      ) ?? [],
  )
  const pastProjects = computed(
    () =>
      p.value?.filter(
        (project) =>
          !currentProjects.value.includes(project) &&
          new Date(project.endDate) < new Date(),
      ) ?? [],
  )
  const futureProjects = computed(
    () =>
      p.value?.filter(
        (project) =>
          !currentProjects.value.includes(project) &&
          new Date(project.startDate) > new Date(),
      ) ?? [],
  )

  return {
    pastProjects,
    futureProjects,
    currentProjects,
  }
}

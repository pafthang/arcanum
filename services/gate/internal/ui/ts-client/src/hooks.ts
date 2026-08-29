import client from './client';

export function useIssues(spaceId: string) {
  const [issues, setIssues] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchIssues = async () => {
    setLoading(true);
    try {
      const res = await client.get(`/spaces/${spaceId}/work/issues`);
      setIssues(res.data);
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchIssues();
  }, [spaceId]);

  return { issues, loading, refetch: fetchIssues };
}

export function useOverview(spaceId: string) {
  const [overview, setOverview] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  const fetchOverview = async () => {
    setLoading(true);
    try {
      const res = await client.get(`/spaces/${spaceId}/work/overview`);
      setOverview(res.data);
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOverview();
  }, [spaceId]);

  return { overview, loading, refetch: fetchOverview };
}

import { createRemnawave } from '@arcanum/ts-client';

export const rw = createRemnawave({ baseUrl: '' });
export const session = rw.session;

import * as generated from '../../bindings/codex-provider-hub/internal/interfaces/wails/app';
import type {
  ModelDTO,
  ProfileDTO,
  ProviderDTO,
  RouteDTO,
} from '../../bindings/codex-provider-hub/internal/interfaces/wails/models';

export type { ModelDTO, ProfileDTO, ProviderDTO, RouteDTO };

export const wailsApi = {
  listProviders: (): Promise<ProviderDTO[] | null> => generated.ListProviders(),
  listModels: (): Promise<ModelDTO[] | null> => generated.ListModels(),
  listModelsByProvider: (providerID: string): Promise<ModelDTO[] | null> => generated.ListModelsByProvider(providerID),
  listRoutes: (): Promise<RouteDTO[] | null> => generated.ListRoutes(),
  listProfiles: (): Promise<ProfileDTO[] | null> => generated.ListProfiles(),
  createProvider: (id: string, name: string, baseURL: string, authRef: string): Promise<ProviderDTO> =>
    generated.CreateProvider(id, name, baseURL, authRef),
  saveProvider: (item: ProviderDTO): Promise<void> => generated.SaveProvider(item),
  deleteProvider: (id: string): Promise<void> => generated.DeleteProvider(id),
  testProvider: (id: string): Promise<void> => generated.TestProvider(id),
  createModel: (providerID: string, id: string, name: string): Promise<ModelDTO> =>
    generated.CreateModel(providerID, id, name),
  saveModel: (item: ModelDTO): Promise<void> => generated.SaveModel(item),
  deleteModel: (providerID: string, id: string): Promise<void> => generated.DeleteModel(providerID, id),
  testModel: (providerID: string, id: string): Promise<void> => generated.TestProviderModel(providerID, id),
  createRoute: (id: string, name: string, providerID: string, modelID: string): Promise<RouteDTO> =>
    generated.CreateRoute(id, name, providerID, modelID),
  saveRoute: (item: RouteDTO): Promise<void> => generated.SaveRoute(item),
  deleteRoute: (id: string): Promise<void> => generated.DeleteRoute(id),
  activateRoute: (routeID: string, profileID: string): Promise<void> => generated.ActivateRoute(routeID, profileID),
  createProfile: (id: string, name: string): Promise<ProfileDTO> => generated.CreateProfile(id, name),
  saveProfile: (item: ProfileDTO): Promise<void> => generated.SaveProfile(item),
  deleteProfile: (id: string): Promise<void> => generated.DeleteProfile(id),
  restoreProfile: (id: string): Promise<void> => generated.RestoreCodexConfig(id),
  codexRunning: (): Promise<boolean> => generated.CodexRunning(),
  startCodex: (): Promise<void> => generated.StartCodex(),
  autostartEnabled: (): Promise<boolean> => generated.AutostartEnabled(),
  setAutostart: (enabled: boolean): Promise<void> => generated.SetAutostart(enabled),
};

export function describeError(error: unknown, fallback = '操作失败'): string {
  if (error instanceof Error && error.message.trim()) return error.message;
  if (typeof error === 'string' && error.trim()) return error;
  return fallback;
}

import{D as i,V as o,M}from"./MeshResources.608317f5.js";import{d as y,k as a,l as P,f as w,g as D,n as t,c as d,p as g,a as s,F as k,x,o as c,e as z}from"./index.f4381a04.js";const I={class:"chart-container mt-16"},S=y({__name:"MainOverviewView",setup(O){const e=x(),n=a(()=>e.getters["config/getMulticlusterStatus"]),u=a(()=>e.getters.getServiceResourcesFetching),l=a(()=>e.getters.getZonesInsightsFetching),r=a(()=>e.getters.getMeshInsightsFetching),_=a(()=>e.getters.getChart("services")),v=a(()=>e.getters.getChart("dataplanes")),f=a(()=>e.getters.getChart("meshes")),p=a(()=>e.getters.getChart("zones")),C=a(()=>e.getters.getChart("zonesCPVersions")),m=a(()=>e.getters.getChart("kumaDPVersions")),V=a(()=>e.getters.getChart("envoyVersions"));P(()=>n.value,function(){h()}),h();function h(){e.dispatch("fetchMeshInsights"),e.dispatch("fetchServices"),e.dispatch("fetchZonesInsights",n.value),n.value&&e.dispatch("fetchTotalClusterCount")}return(Z,F)=>(c(),w(k,null,[D("div",I,[t(n)?(c(),d(i,{key:0,class:"chart chart-1/2 chart-offset-left-1/6",title:{singular:"Zone",plural:"Zones"},data:t(p).data,url:{name:"zones"},"is-loading":t(l)},null,8,["data","is-loading"])):g("",!0),t(n)?(c(),d(o,{key:1,class:"chart chart-1/2 chart-offset-right-1/6",title:"ZONE CP",data:t(C).data,url:{name:"zones"},"is-loading":t(l)},null,8,["data","is-loading"])):g("",!0),s(i,{class:"chart chart-1/3",title:{singular:"Mesh",plural:"Meshes"},data:t(f).data,"is-loading":t(r)},null,8,["data","is-loading"]),s(i,{class:"chart chart-1/3",title:{singular:"Service",plural:"Services"},data:t(_).data,"is-loading":t(u),"save-chart":""},null,8,["data","is-loading"]),s(i,{class:"chart chart-1/3",title:{singular:"DP Proxy",plural:"DP Proxies"},data:t(v).data,"is-loading":t(r)},null,8,["data","is-loading"]),s(o,{class:"chart chart-1/2 chart-offset-left-1/6",title:"KUMA DP",data:t(m).data,"is-loading":t(r)},null,8,["data","is-loading"]),s(o,{class:"chart chart-1/2 chart-offset-right-1/6",title:"ENVOY",data:t(V).data,"is-loading":t(r),"display-am-charts-logo":""},null,8,["data","is-loading"])]),s(M,{class:"mt-8"})],64))}});const E=z(S,[["__scopeId","data-v-67a6be61"]]);export{E as default};

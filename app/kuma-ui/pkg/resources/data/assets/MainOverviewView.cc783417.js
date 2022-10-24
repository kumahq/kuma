import{_ as r,a as o,M as V}from"./MeshResources.f2d6b4f7.js";import{d as y,f as a,g as P,o as c,k as w,l as k,u as t,c as d,y as g,a as s,F as x,q as z,C as I}from"./index.ea0d4a24.js";const S={class:"chart-container mt-16"},D=y({__name:"MainOverviewView",setup(O){const e=z(),n=a(()=>e.getters["config/getMulticlusterStatus"]),u=a(()=>e.getters.getServiceResourcesFetching),l=a(()=>e.getters.getZonesInsightsFetching),i=a(()=>e.getters.getMeshInsightsFetching),_=a(()=>e.getters.getChart("services")),f=a(()=>e.getters.getChart("dataplanes")),v=a(()=>e.getters.getChart("meshes")),p=a(()=>e.getters.getChart("zones")),m=a(()=>e.getters.getChart("zonesCPVersions")),C=a(()=>e.getters.getChart("kumaDPVersions")),M=a(()=>e.getters.getChart("envoyVersions"));P(()=>n.value,function(){h()}),h();function h(){e.dispatch("fetchMeshInsights"),e.dispatch("fetchServices"),e.dispatch("fetchZonesInsights",n.value),n.value&&e.dispatch("fetchTotalClusterCount")}return(Z,F)=>(c(),w(x,null,[k("div",S,[t(n)?(c(),d(r,{key:0,class:"chart chart-1/2 chart-offset-left-1/6",title:{singular:"Zone",plural:"Zones"},data:t(p).data,url:{name:"zones"},"is-loading":t(l)},null,8,["data","is-loading"])):g("",!0),t(n)?(c(),d(o,{key:1,class:"chart chart-1/2 chart-offset-right-1/6",title:"ZONE CP",data:t(m).data,url:{name:"zones"},"is-loading":t(l)},null,8,["data","is-loading"])):g("",!0),s(r,{class:"chart chart-1/3",title:{singular:"Mesh",plural:"Meshes"},data:t(v).data,"is-loading":t(i)},null,8,["data","is-loading"]),s(r,{class:"chart chart-1/3",title:{singular:"Service",plural:"Services"},data:t(_).data,"is-loading":t(u),"save-chart":""},null,8,["data","is-loading"]),s(r,{class:"chart chart-1/3",title:{singular:"DP Proxy",plural:"DP Proxies"},data:t(f).data,"is-loading":t(i)},null,8,["data","is-loading"]),s(o,{class:"chart chart-1/2 chart-offset-left-1/6",title:"KUMA DP",data:t(C).data,"is-loading":t(i)},null,8,["data","is-loading"]),s(o,{class:"chart chart-1/2 chart-offset-right-1/6",title:"ENVOY",data:t(M).data,"is-loading":t(i),"display-am-charts-logo":""},null,8,["data","is-loading"])]),s(V,{class:"mt-8"})],64))}});const E=I(D,[["__scopeId","data-v-67a6be61"]]);export{E as default};

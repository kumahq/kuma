import{d as C,e as i,x as D,i as j,o as a,j as l,l as t,b as n,a as S,w as I,t as u,u as r,c as k,z as d,F as g,C as N,D as O,E as V}from"./index.0cb244cf.js";import{S as E}from"./StatusBadge.291bc99c.js";import{T as P}from"./TagList.9a369e5a.js";import{_ as $}from"./YamlView.vue_vue_type_script_setup_true_lang.a2d41d2e.js";const _=c=>(N("data-v-312508fb"),c=c(),O(),c),q={class:"entity-summary entity-section-list"},F={class:"block-list"},L={class:"entity-title"},z={class:"definition"},A=_(()=>t("span",null,"Mesh:",-1)),M={class:"definition"},R=_(()=>t("span",null,"Address:",-1)),G={key:0,class:"definition"},H=_(()=>t("span",null,"TLS:",-1)),J={key:1,class:"definition"},K=_(()=>t("span",null,"Data plane proxies:",-1)),Q={key:0},U=_(()=>t("h2",null,"Tags",-1)),W={class:"config-section"},X=C({__name:"ServiceSummary",props:{service:{type:Object,required:!0},externalService:{type:Object,required:!1,default:null}},setup(c){const e=c,b=i(()=>({name:"service-detail-view",params:{service:e.service.name,mesh:e.service.mesh}})),p=i(()=>{var s;return e.service.serviceType==="external"&&e.externalService!==null?e.externalService.networking.address:(s=e.service.addressPort)!=null?s:null}),m=i(()=>{var s;return e.service.serviceType==="external"&&e.externalService!==null?(s=e.externalService.networking.tls)!=null&&s.enabled?"Enabled":"Disabled":null}),h=i(()=>{var s,o,v,y;if(e.service.serviceType==="external")return null;{const w=(o=(s=e.service.dataplanes)==null?void 0:s.online)!=null?o:0,B=(y=(v=e.service.dataplanes)==null?void 0:v.total)!=null?y:0;return`${w} online / ${B} total`}}),f=i(()=>{var s;return e.service.serviceType==="external"?null:(s=e.service.status)!=null?s:null}),x=i(()=>e.service.serviceType==="external"&&e.externalService!==null?Object.entries(e.externalService.tags).map(([s,o])=>({label:s,value:o})):[]),T=i(()=>{var s;return D((s=e.externalService)!=null?s:e.service)});return(s,o)=>{const v=j("router-link");return a(),l("div",q,[t("section",null,[t("div",F,[t("div",null,[t("h1",L,[t("span",null,[n(`
              Service:

              `),S(v,{to:r(b)},{default:I(()=>[n(u(e.service.name),1)]),_:1},8,["to"])]),n(),r(f)?(a(),k(E,{key:0,status:r(f)},null,8,["status"])):d("",!0)]),n(),t("div",z,[A,n(),t("span",null,u(e.service.mesh),1)]),n(),t("div",M,[R,n(),t("span",null,[r(p)!==null?(a(),l(g,{key:0},[n(u(r(p)),1)],64)):(a(),l(g,{key:1},[n("\u2014")],64))])]),n(),r(m)!==null?(a(),l("div",G,[H,n(),t("span",null,u(r(m)),1)])):d("",!0),n(),r(h)!==null?(a(),l("div",J,[K,n(),t("span",null,u(r(h)),1)])):d("",!0)]),n(),r(x).length>0?(a(),l("div",Q,[U,n(),S(P,{tags:r(x)},null,8,["tags"])])):d("",!0)])]),n(),t("section",W,[e.service.serviceType==="external"?(a(),k($,{key:0,id:"code-block-service",content:r(T),"is-searchable":"","code-max-height":"250px"},null,8,["content"])):d("",!0)])])}}});const te=V(X,[["__scopeId","data-v-312508fb"]]);export{te as S};

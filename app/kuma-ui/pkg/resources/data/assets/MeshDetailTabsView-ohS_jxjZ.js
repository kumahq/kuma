import{_ as u}from"./NavTabs.vue_vue_type_script_setup_true_lang-7vT9lZHC.js";import{d,a as s,o as f,b as h,w as t,V as w,e as a,m as V,f as n,t as v,H as x,W as R}from"./index-8scsg5Gp.js";const $=d({__name:"MeshDetailTabsView",setup(b){return(T,k)=>{const m=s("RouteTitle"),c=s("RouterLink"),l=s("RouterView"),_=s("AppView"),p=s("RouteView");return f(),h(p,{name:"mesh-detail-tabs-view",params:{mesh:""}},{default:t(({route:o,t:i})=>[a(_,null,{title:t(()=>[V("h1",null,[a(w,{text:o.params.mesh},{default:t(()=>[a(m,{title:i("meshes.routes.item.title",{name:o.params.mesh})},null,8,["title"])]),_:2},1032,["text"])])]),default:t(()=>{var r;return[n(),a(u,{"active-route-name":(r=o.active)==null?void 0:r.name,"data-testid":"mesh-tabs"},R({_:2},[x(o.children.filter(({name:e})=>e!=="external-service-list-view"),({name:e})=>({name:`${e}`,fn:t(()=>[a(c,{to:{name:e},"data-testid":`${e}-tab`},{default:t(()=>[n(v(i(`meshes.routes.item.navigation.${e}`)),1)]),_:2},1032,["to","data-testid"])])}))]),1032,["active-route-name"]),n(),a(l)]}),_:2},1024)]),_:1})}}});export{$ as default};

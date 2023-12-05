import{L as D}from"./LoadingBox-e89ad34d.js";import{O as T,a as O,b as V}from"./OnboardingPage-d5a6f7d8.js";import{S as B}from"./StatusBadge-e45862f2.js";import{g as R}from"./index-3d038f44.js";import{d as S,u as N,m as k,T as P,a as u,o,b as v,w as t,e as s,f as c,c as p,F as x,J as E,t as f,p as b,_ as F}from"./index-f56c27ab.js";const L={key:0,class:"status-loading-box mb-4"},C={key:1},K={class:"mb-4"},$=S({__name:"OnboardingDataplanesView",setup(H){const h=N(),A=[{label:"Mesh",key:"mesh"},{label:"Name",key:"name"},{label:"Status",key:"status"}],a=k({total:0,data:[]}),_=k(null);P(function(){w()}),y();function w(){_.value!==null&&window.clearTimeout(_.value)}async function y(){let i=!1;const m=[];try{const{items:n}=await h.getAllDataplanes({size:10});if(Array.isArray(n)&&n.length>0)for(const g of n){const{name:r,mesh:d}=g,l=await h.getDataplaneOverviewFromMesh({mesh:d,name:r}),{status:e}=R(l);e==="offline"&&(i=!0),m.push({status:e,name:r,mesh:d})}else i=!0}catch(n){console.error(n)}a.value.data=m,a.value.total=a.value.data.length,i&&(w(),_.value=window.setTimeout(y,1e3))}return(i,m)=>{const n=u("RouteTitle"),g=u("KTable"),r=u("AppView"),d=u("RouteView");return o(),v(d,{name:"onboarding-dataplanes-view"},{default:t(({t:l})=>[s(n,{title:l("onboarding.routes.dataplanes-overview.title"),render:!1},null,8,["title"]),c(),s(r,null,{default:t(()=>[s(T,null,{header:t(()=>[(o(!0),p(x,null,E([a.value.data.length>0?"success":"waiting"],e=>(o(),v(O,{key:e,"data-testid":`state-${e}`},{title:t(()=>[c(f(l(`onboarding.routes.dataplanes-overview.header.${e}.title`)),1)]),description:t(()=>[b("p",null,f(l(`onboarding.routes.dataplanes-overview.header.${e}.description`)),1)]),_:2},1032,["data-testid"]))),128))]),content:t(()=>[a.value.data.length===0?(o(),p("div",L,[s(D)])):(o(),p("div",C,[b("p",K,[b("b",null,"Found "+f(a.value.data.length)+" DPPs:",1)]),c(),s(g,{class:"mb-4",fetcher:()=>a.value,headers:A,"disable-pagination":""},{status:t(({rowValue:e})=>[e?(o(),v(B,{key:0,status:e},null,8,["status"])):(o(),p(x,{key:1},[c(`
                  —
                `)],64))]),_:1},8,["fetcher"])]))]),navigation:t(()=>[s(V,{"next-step":"onboarding-completed-view","previous-step":"onboarding-add-new-services-code-view","should-allow-next":a.value.data.length>0},null,8,["should-allow-next"])]),_:2},1024)]),_:2},1024)]),_:1})}}});const j=F($,[["__scopeId","data-v-0917d04b"]]);export{j as default};

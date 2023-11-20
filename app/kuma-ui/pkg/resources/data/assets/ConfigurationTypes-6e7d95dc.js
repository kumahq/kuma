import{O as h,a as x,b as C}from"./OnboardingPage-f6a09138.js";import{d as G,k as R,P as T,Q as w,R as P,m as K,O,a as r,o as u,b as p,w as e,e as o,f as n,p as d,H as k,_ as M}from"./index-203d56a2.js";const N={class:"graph-list mb-6"},U={class:"radio-button-group"},A=G({__name:"ConfigurationTypes",setup(B){const i=R(),m=T(),_=w(),c={postgres:P(),memory:_,kubernetes:m},t=K(i("KUMA_STORE_TYPE")),g=O(()=>c[t.value]);return(z,a)=>{const v=r("RouteTitle"),l=r("KRadio"),b=r("AppView"),f=r("RouteView");return u(),p(f,{name:"onboarding-configuration-types"},{default:e(({can:y,t:V})=>[o(v,{title:V("onboarding.routes.configuration-types.title"),render:!1},null,8,["title"]),n(),o(b,null,{default:e(()=>[o(h,{"with-image":""},{header:e(()=>[o(x,null,{title:e(()=>[n(`
              Learn about configuration storage
            `)]),_:1})]),content:e(()=>[d("div",N,[(u(),p(k(g.value)))]),n(),d("div",U,[o(l,{modelValue:t.value,"onUpdate:modelValue":a[0]||(a[0]=s=>t.value=s),name:"deployment","selected-value":"kubernetes"},{default:e(()=>[n(`
              Kubernetes
            `)]),_:1},8,["modelValue"]),n(),o(l,{modelValue:t.value,"onUpdate:modelValue":a[1]||(a[1]=s=>t.value=s),name:"deployment","selected-value":"postgres"},{default:e(()=>[n(`
              Postgres
            `)]),_:1},8,["modelValue"]),n(),o(l,{modelValue:t.value,"onUpdate:modelValue":a[2]||(a[2]=s=>t.value=s),name:"deployment","selected-value":"memory"},{default:e(()=>[n(`
              Memory
            `)]),_:1},8,["modelValue"])])]),navigation:e(()=>[o(C,{"next-step":y("use zones")?"onboarding-multi-zone":"onboarding-create-mesh","previous-step":"onboarding-deployment-types"},null,8,["next-step"])]),_:2},1024)]),_:2},1024)]),_:1})}}});const I=M(A,[["__scopeId","data-v-7be26533"]]);export{I as default};

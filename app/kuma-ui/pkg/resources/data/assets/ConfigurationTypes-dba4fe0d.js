import{O as h,a as x,b as C}from"./OnboardingPage-45fd8d0f.js";import{d as G,M as T,N as w,O as P,P as R,y as K,h as M,r,o as u,i as d,w as e,j as o,n as t,p,q as O,t as N}from"./index-24ab6634.js";const k={class:"graph-list mb-6"},U={class:"radio-button-group"},A=G({__name:"ConfigurationTypes",setup(B){const i=T(),m=w(),c=P(),_={postgres:R(),memory:c,kubernetes:m},n=K(i("KUMA_STORE_TYPE")),g=M(()=>_[n.value]);return(z,a)=>{const v=r("RouteTitle"),l=r("KRadio"),y=r("AppView"),f=r("RouteView");return u(),d(f,{name:"onboarding-configuration-types"},{default:e(({can:b,t:V})=>[o(v,{title:V("onboarding.routes.configuration-types.title")},null,8,["title"]),t(),o(y,null,{default:e(()=>[o(h,{"with-image":""},{header:e(()=>[o(x,null,{title:e(()=>[t(`
              Learn about configuration storage
            `)]),_:1})]),content:e(()=>[p("div",k,[(u(),d(O(g.value)))]),t(),p("div",U,[o(l,{modelValue:n.value,"onUpdate:modelValue":a[0]||(a[0]=s=>n.value=s),name:"deployment","selected-value":"kubernetes"},{default:e(()=>[t(`
              Kubernetes
            `)]),_:1},8,["modelValue"]),t(),o(l,{modelValue:n.value,"onUpdate:modelValue":a[1]||(a[1]=s=>n.value=s),name:"deployment","selected-value":"postgres"},{default:e(()=>[t(`
              Postgres
            `)]),_:1},8,["modelValue"]),t(),o(l,{modelValue:n.value,"onUpdate:modelValue":a[2]||(a[2]=s=>n.value=s),name:"deployment","selected-value":"memory"},{default:e(()=>[t(`
              Memory
            `)]),_:1},8,["modelValue"])])]),navigation:e(()=>[o(C,{"next-step":b("use zones")?"onboarding-multi-zone":"onboarding-create-mesh","previous-step":"onboarding-deployment-types"},null,8,["next-step"])]),_:2},1024)]),_:2},1024)]),_:1})}}});const D=N(A,[["__scopeId","data-v-d26eecda"]]);export{D as default};

import{C as x}from"./CodeBlock-3HK16lYA.js";import{d as V,a as r,o as n,b as B,w as t,Z as s,f as a,e,c as l,t as p,W as _,q as h,F as T,m as i,p as z,H as D}from"./index-i9Uphcre.js";import{T as y}from"./TagList-lTdCg45Q.js";import{t as N}from"./toYaml-sPaYOD3i.js";const A={class:"stack"},P={key:0,class:"stack-with-borders"},$={key:1,class:"stack-with-borders"},L={key:2},S=i("h3",null,"Rules",-1),R={class:"stack mt-4"},F={class:"mt-1"},O=V({__name:"ConnectionInboundSummaryOverviewView",props:{inbound:{},gateway:{}},setup(b){const o=b;return(H,K)=>{const m=r("KBadge"),g=r("DataCollection"),f=r("DataLoader"),v=r("AppView"),w=r("RouteView");return n(),B(w,{name:"connection-inbound-summary-overview-view",params:{mesh:"",dataPlane:"",service:""}},{default:t(({route:c,t:k})=>[e(v,null,{default:t(()=>[i("div",A,[o.gateway?(n(),l("div",P,[e(s,{layout:"horizontal"},{title:t(()=>[a(`
              Tags
            `)]),body:t(()=>[e(y,{tags:o.gateway.tags,alignment:"right"},null,8,["tags"])]),_:1})])):o.inbound?(n(),l("div",$,[e(s,{layout:"horizontal"},{title:t(()=>[a(`
              Tags
            `)]),body:t(()=>[e(y,{tags:o.inbound.tags,alignment:"right"},null,8,["tags"])]),_:1}),a(),e(s,{layout:"horizontal"},{title:t(()=>[a(`
              Status
            `)]),body:t(()=>[e(m,{appearance:o.inbound.health.ready?"success":"danger"},{default:t(()=>[a(p(o.inbound.health.ready?"Healthy":"Unhealthy"),1)]),_:1},8,["appearance"])]),_:1}),a(),e(s,{layout:"horizontal"},{title:t(()=>[a(`
              Protocol
            `)]),body:t(()=>[e(m,{appearance:"info"},{default:t(()=>[a(p(k(`http.api.value.${o.inbound.protocol}`)),1)]),_:2},1024)]),_:2},1024),a(),e(s,{layout:"horizontal"},{title:t(()=>[a(`
              Address
            `)]),body:t(()=>[e(_,{text:`${o.inbound.addressPort}`},null,8,["text"])]),_:1}),a(),e(s,{layout:"horizontal"},{title:t(()=>[a(`
              Service Address
            `)]),body:t(()=>[e(_,{text:`${o.inbound.serviceAddressPort}`},null,8,["text"])]),_:1})])):h("",!0),a(),o.inbound?(n(),l("div",L,[S,a(),e(f,{src:`/meshes/${c.params.mesh}/dataplanes/${c.params.dataPlane}/rules`},{default:t(({data:C})=>[e(g,{predicate:d=>d.ruleType==="from"&&Number(d.inbound.port)===Number(c.params.service.substring(1)),items:C.rules},{default:t(({items:d})=>[i("dl",R,[(n(!0),l(T,null,D(d,u=>(n(),l("div",{key:u},[i("dt",null,p(u.type),1),a(),i("dd",F,[e(x,{code:z(N)(u.config),language:"yaml","show-copy-button":!1},null,8,["code"])])]))),128))])]),_:2},1032,["predicate","items"])]),_:2},1032,["src"])])):h("",!0)])]),_:2},1024)]),_:1})}}});export{O as default};

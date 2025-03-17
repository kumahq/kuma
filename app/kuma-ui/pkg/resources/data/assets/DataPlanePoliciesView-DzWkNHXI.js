import{a as K,A as z}from"./AccordionList-MqsYTHMd.js";import{d as I,r as k,o as e,c as s,s as u,e as l,F as i,v as _,t as p,b as o,w as a,q as j,m as g,T as Y,_ as T,p as O,a8 as H,k as J,C as Q,n as U,ac as W,ad as Z,K as D}from"./index-Bi3CXAeE.js";import{S as E}from"./SummaryView-DSsmf-Dn.js";import{P as M}from"./PolicyTypeTag-mg1KvcGa.js";import{T as q}from"./TagList-DYjb-hhI.js";import{R as ee}from"./RuleMatchers-ShEVNSLD.js";const te={class:"policies-list"},se={class:"mesh-gateway-policy-list"},ae={key:0},le={class:"dataplane-policy-header"},ne={key:0,class:"badge-list"},oe={class:"mt-1"},ie=I({__name:"BuiltinGatewayPolicies",props:{gatewayDataplane:{},types:{}},setup(V){const w=V;return(A,t)=>{const L=k("XAction"),C=k("XBadge");return e(),s("div",te,[u("div",se,[t[11]||(t[11]=u("h3",{class:"mb-2"},`
        Gateway policies
      `,-1)),t[12]||(t[12]=l()),A.gatewayDataplane.routePolicies.length>0?(e(),s("ul",ae,[(e(!0),s(i,null,_(A.gatewayDataplane.routePolicies,(m,c)=>{var r;return e(),s("li",{key:c},[u("span",null,p(m.type),1),t[0]||(t[0]=l(`:

          `)),o(L,{to:{name:"policy-detail-view",params:{mesh:m.mesh,policyPath:((r=w.types[m.type])==null?void 0:r.path)??"",policy:m.name}}},{default:a(()=>[l(p(m.name),1)]),_:2},1032,["to"])])}),128))])):j("",!0),t[13]||(t[13]=l()),t[14]||(t[14]=u("h3",{class:"mt-6 mb-2"},`
        Listeners
      `,-1)),t[15]||(t[15]=l()),u("div",null,[(e(!0),s(i,null,_(A.gatewayDataplane.listenerEntries,(m,c)=>(e(),s("div",{key:c},[u("div",null,[u("div",null,[t[1]||(t[1]=u("b",null,"Host",-1)),l(": "+p(m.hostName)+":"+p(m.port)+" ("+p(m.protocol)+`)
            `,1)]),t[10]||(t[10]=l()),m.routeEntries.length>0?(e(),s(i,{key:0},[t[8]||(t[8]=u("h4",{class:"mt-2 mb-2"},`
                Routes
              `,-1)),t[9]||(t[9]=l()),o(z,{"initially-open":[],"multiple-open":""},{default:a(()=>[(e(!0),s(i,null,_(m.routeEntries,(r,P)=>(e(),g(K,{key:P},Y({"accordion-header":a(()=>{var n;return[u("div",le,[u("div",null,[u("div",null,[t[2]||(t[2]=u("b",null,"Route",-1)),t[3]||(t[3]=l(": ")),o(L,{to:{name:"policy-detail-view",params:{mesh:r.route.mesh,policyPath:((n=w.types[r.route.type])==null?void 0:n.path)??"",policy:r.route.name}}},{default:a(()=>[l(p(r.route.name),1)]),_:2},1032,["to"])]),t[5]||(t[5]=l()),u("div",null,[t[4]||(t[4]=u("b",null,"Service",-1)),l(": "+p(r.service),1)])]),t[6]||(t[6]=l()),r.origins.length>0?(e(),s("div",ne,[(e(!0),s(i,null,_(r.origins,(y,h)=>(e(),g(C,{key:`${c}-${h}`},{default:a(()=>[l(p(y.type),1)]),_:2},1024))),128))])):j("",!0)])]}),_:2},[r.origins.length>0?{name:"accordion-content",fn:a(()=>[u("ul",oe,[(e(!0),s(i,null,_(r.origins,(n,y)=>{var h;return e(),s("li",{key:`${c}-${y}`},[l(p(n.type)+`:

                        `,1),o(L,{to:{name:"policy-detail-view",params:{mesh:n.mesh,policyPath:((h=w.types[n.type])==null?void 0:h.path)??"",policy:n.name}}},{default:a(()=>[l(p(n.name),1)]),_:2},1032,["to"])])}),128))])]),key:"0"}:void 0]),1024))),128))]),_:2},1024)],64)):j("",!0)])]))),128))])])])}}}),re=T(ie,[["__scopeId","data-v-a5c76432"]]),ue={class:"policy-type-heading"},pe={class:"policy-list"},de={key:0},me=I({__name:"PolicyTypeEntryList",props:{items:{},types:{}},setup(V){const w=V;function A({headerKey:t}){return{class:`cell-${t}`}}return(t,L)=>{const C=k("XAction"),m=k("XCodeBlock"),c=k("KTable");return e(),g(z,{"initially-open":[],"multiple-open":""},{default:a(()=>[(e(!0),s(i,null,_(t.items,(r,P)=>(e(),g(K,{key:P},{"accordion-header":a(()=>[u("h3",ue,[o(M,{"policy-type":r.type},{default:a(()=>[l(p(r.type)+" ("+p(r.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":a(()=>[u("div",pe,[o(c,{class:"policy-type-table",fetcher:()=>({data:r.connections,total:r.connections.length}),headers:[{label:"From",key:"sourceTags"},{label:"To",key:"destinationTags"},{label:"On",key:"name"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}],"cell-attrs":A,"disable-pagination":"","is-clickable":""},{sourceTags:a(({row:n})=>[n.sourceTags.length>0?(e(),g(q,{key:0,class:"tag-list","should-truncate":"",tags:n.sourceTags},null,8,["tags"])):(e(),s(i,{key:1},[l(`
                —
              `)],64))]),destinationTags:a(({row:n})=>[n.destinationTags.length>0?(e(),g(q,{key:0,class:"tag-list","should-truncate":"",tags:n.destinationTags},null,8,["tags"])):(e(),s(i,{key:1},[l(`
                —
              `)],64))]),name:a(({row:n})=>[n.name!==null?(e(),s(i,{key:0},[l(p(n.name),1)],64)):(e(),s(i,{key:1},[l(`
                —
              `)],64))]),origins:a(({row:n})=>[n.origins.length>0?(e(),s("ul",de,[(e(!0),s(i,null,_(n.origins,(y,h)=>{var S;return e(),s("li",{key:`${P}-${h}`},[o(C,{to:{name:"policy-detail-view",params:{mesh:y.mesh,policyPath:((S=w.types[y.type])==null?void 0:S.path)??"",policy:y.name}}},{default:a(()=>[l(p(y.name),1)]),_:2},1032,["to"])])}),128))])):(e(),s(i,{key:1},[l(`
                —
              `)],64))]),config:a(({row:n})=>[n.config?(e(),g(m,{key:0,code:O(H).stringify(n.config),language:"yaml","show-copy-button":!1},null,8,["code"])):(e(),s(i,{key:1},[l(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}}),ce=T(me,[["__scopeId","data-v-3dd03e37"]]),ye={class:"policy-type-heading"},fe={key:0,class:"matcher"},_e={key:1},ge={key:0},ke=I({__name:"RuleList",props:{rules:{},types:{}},setup(V){const{t:w}=J(),A=V;return(t,L)=>{const C=k("XAction"),m=k("XCodeBlock");return e(),g(z,{"initially-open":[],"multiple-open":""},{default:a(()=>[(e(!0),s(i,null,_([A.rules.reduce((c,r)=>(typeof c[r.type]>"u"&&(c[r.type]=[]),c[r.type].push(r),c),{})],c=>(e(),s(i,{key:c},[(e(!0),s(i,null,_(c,(r,P)=>(e(),g(K,{key:P},{"accordion-header":a(()=>[u("h3",ye,[o(M,{"policy-type":P},{default:a(()=>[l(p(P),1)]),_:2},1032,["policy-type"])])]),"accordion-content":a(()=>[(e(!0),s(i,null,_([r.some(n=>n.matchers.length>0)],n=>(e(),s("div",{key:n,class:"policy-list"},[o(Q,{class:U(["policy-type-table",{"has-matchers":n}]),items:r,headers:[...n?[{label:"Matchers",key:"matchers"}]:[],{label:"Origin policies",key:"origins"},{label:"Conf",key:"config"}]},{matchers:a(({row:y})=>[y.matchers.length>0?(e(),s("span",fe,[o(ee,{items:y.matchers},null,8,["items"])])):(e(),s("i",_e,p(O(w)("data-planes.routes.item.matches_everything")),1))]),origins:a(({row:y})=>[y.origins.length>0?(e(),s("ul",ge,[(e(!0),s(i,null,_(y.origins,(h,S)=>{var v;return e(),s("li",{key:`${P}-${S}`},[o(C,{to:{name:"data-plane-policy-summary-view",params:{mesh:h.mesh,policyPath:((v=A.types[h.type])==null?void 0:v.path)??"",policy:h.name}}},{default:a(()=>[l(p(h.name),1)]),_:2},1032,["to"])])}),128))])):(e(),s(i,{key:1},[l(p(O(w)("common.collection.none")),1)],64))]),config:a(({row:y})=>[Object.keys(y.raw).length>0?(e(),g(m,{key:0,code:O(H).stringify(y.raw),language:"yaml","show-copy-button":!1},null,8,["code"])):(e(),s(i,{key:1},[l(p(O(w)("common.collection.none")),1)],64))]),_:2},1032,["class","items","headers"])]))),128))]),_:2},1024))),128))],64))),128))]),_:1})}}}),G=T(ke,[["__scopeId","data-v-8b58b2bd"]]),he={class:"mb-2"},be={class:"mb-2"},ve={key:0},Re=I({__name:"DataPlanePoliciesView",props:{data:{}},setup(V){const w=V;return(A,t)=>{const L=k("RouteTitle"),C=k("XCard"),m=k("DataCollection"),c=k("DataLoader"),r=k("RouterView"),P=k("DataSource"),n=k("AppView"),y=k("RouteView");return e(),g(y,{name:"data-plane-policies-view",params:{mesh:"",proxy:""}},{default:a(({uri:h,can:S,route:v,t:R})=>[o(L,{render:!1,title:R("data-planes.routes.item.navigation.data-plane-policies-view")},null,8,["title"]),t[11]||(t[11]=l()),o(n,null,{default:a(()=>[o(P,{src:h(O(W),"/policy-types",{})},{default:a(({data:X,error:F})=>[(e(!0),s(i,null,_([((X==null?void 0:X.policyTypes)??[]).reduce(($,b)=>Object.assign($,{[b.name]:b}),{})],$=>(e(),s(i,{key:typeof $},[o(c,{src:h(O(Z),"/meshes/:mesh/rules/for/:dataplane",{mesh:v.params.mesh,dataplane:v.params.proxy}),data:[X],errors:[F]},{default:a(({data:b})=>[o(m,{items:b.rules},{default:a(()=>[(e(),s(i,null,_(["proxy","to"],f=>o(m,{key:f,items:b.rules,predicate:d=>d.ruleType===f,comparator:(d,B)=>d.type.localeCompare(B.type),empty:!1},{default:a(({items:d})=>[o(C,null,{default:a(()=>[u("h3",null,p(R(`data-planes.routes.item.rules.${f}`)),1),t[0]||(t[0]=l()),o(G,{class:"mt-2",rules:d,types:$,"data-testid":`${f}-rule-list`},null,8,["rules","types","data-testid"])]),_:2},1024)]),_:2},1032,["items","predicate","comparator"])),64)),t[5]||(t[5]=l()),o(m,{items:b.rules,predicate:f=>{var d;return f.ruleType==="from"&&!((d=$[f.type])!=null&&d.policy.isFromAsRules)},comparator:(f,d)=>f.type.localeCompare(d.type),empty:!1},{default:a(({items:f})=>[o(C,null,{default:a(()=>[u("h3",he,p(R("data-planes.routes.item.rules.from")),1),t[2]||(t[2]=l()),(e(!0),s(i,null,_([Object.groupBy(f,d=>d.inbound.port)],d=>(e(),s(i,{key:d},[(e(!0),s(i,null,_(Object.entries(d).sort(([B],[N])=>Number(N)-Number(B)),([B,N],x)=>(e(),s("div",{key:x},[u("h4",null,p(R("data-planes.routes.item.port",{port:B})),1),t[1]||(t[1]=l()),o(G,{class:"mt-2",rules:N,types:$,"data-testid":`from-rule-list-${x}`},null,8,["rules","types","data-testid"])]))),128))],64))),128))]),_:2},1024)]),_:2},1032,["items","predicate","comparator"]),t[6]||(t[6]=l()),o(m,{items:b.inboundRules,comparator:(f,d)=>f.type.localeCompare(d.type),empty:!1},{default:a(({items:f})=>[o(C,null,{default:a(()=>[u("h3",be,p(R("data-planes.routes.item.rules.inbound")),1),t[4]||(t[4]=l()),(e(!0),s(i,null,_([Object.groupBy(f,d=>d.inbound.port)],d=>(e(),s(i,{key:d},[(e(!0),s(i,null,_(Object.entries(d).sort(([B],[N])=>Number(N)-Number(B)),([B,N],x)=>(e(),s("div",{key:x},[u("h4",null,p(R("data-planes.routes.item.port",{port:B})),1),t[3]||(t[3]=l()),o(G,{class:"mt-2",rules:N,types:$,"data-testid":`inbound-rule-list-${x}`},null,8,["rules","types","data-testid"])]))),128))],64))),128))]),_:2},1024)]),_:2},1032,["items","comparator"])]),_:2},1032,["items"])]),_:2},1032,["src","data","errors"]),t[9]||(t[9]=l()),S("use zones")?j("",!0):(e(),s("div",ve,[w.data.dataplaneType==="builtin"?(e(),g(c,{key:0,src:`/meshes/${v.params.mesh}/dataplanes/${v.params.proxy}/gateway-dataplane-policies`,data:[X],errors:[F]},{default:a(({data:b})=>[b?(e(),g(m,{key:0,items:b.routePolicies,empty:!1},{default:a(()=>[u("h3",null,p(R("data-planes.routes.item.legacy_policies")),1),t[7]||(t[7]=l()),o(C,{class:"mt-4"},{default:a(()=>[o(re,{types:$,"gateway-dataplane":b,"data-testid":"builtin-gateway-dataplane-policies"},null,8,["types","gateway-dataplane"])]),_:2},1024)]),_:2},1032,["items"])):j("",!0)]),_:2},1032,["src","data","errors"])):(e(),g(c,{key:1,src:`/meshes/${v.params.mesh}/dataplanes/${v.params.proxy}/sidecar-dataplane-policies`,data:[X],errors:[F]},{default:a(({data:b})=>[o(m,{predicate:f=>{var d;return((d=$[f.type])==null?void 0:d.policy.isTargetRef)===!1},items:b.policyTypeEntries,empty:!1},{default:a(({items:f})=>[u("h3",null,p(R("data-planes.routes.item.legacy_policies")),1),t[8]||(t[8]=l()),o(C,{class:"mt-4"},{default:a(()=>[o(ce,{items:f,types:$,"data-testid":"sidecar-dataplane-policies"},null,8,["items","types"])]),_:2},1024)]),_:2},1032,["predicate","items"])]),_:2},1032,["src","data","errors"]))]))],64))),128)),t[10]||(t[10]=l()),o(r,null,{default:a(({Component:$})=>[v.child()&&X?(e(),g(E,{key:0,onClose:b=>v.replace({name:"data-plane-policies-view",params:{mesh:v.params.mesh,proxy:v.params.proxy}})},{default:a(()=>[(e(),g(D($),{"policy-types":X.policyTypes},null,8,["policy-types"]))]),_:2},1032,["onClose"])):j("",!0)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});export{Re as default};

import{d as T,F as L,e as w,o as s,c as a,m as C,w as c,a as _,r as S,b as e,p,k as n,l as y,K as H,t as i,q as B,H as u,J as v,j as I,n as E,ao as N,an as V}from"./index-CjjKwNo4.js";import{T as X}from"./TagList-pzFi9naC.js";import{R as A}from"./RuleMatchers-C94XW3d9.js";function F(k,l){for(const m of l)if(m.not&&k[m.key]===m.value||!m.not&&k[m.key]!==m.value)return!1;return!0}const D="data:image/svg+xml,%3csvg%20width='11'%20height='12'%20viewBox='0%200%2011%2012'%20fill='none'%20xmlns='http://www.w3.org/2000/svg'%3e%3cpath%20d='M5.66661%202.66667C5.8555%202.66667%206.01384%202.60278%206.14161%202.475C6.26939%202.34722%206.33328%202.18889%206.33328%202C6.33328%201.81111%206.26939%201.65278%206.14161%201.525C6.01384%201.39722%205.8555%201.33333%205.66661%201.33333C5.47773%201.33333%205.31939%201.39722%205.19161%201.525C5.06384%201.65278%204.99995%201.81111%204.99995%202C4.99995%202.18889%205.06384%202.34722%205.19161%202.475C5.31939%202.60278%205.47773%202.66667%205.66661%202.66667ZM7.54995%202.66667H8.71661C9.04995%202.66667%209.33884%202.77778%209.58328%203C9.82773%203.22222%209.97773%203.49444%2010.0333%203.81667L10.9833%2010.4833C11.0388%2010.8833%2010.9361%2011.2361%2010.6749%2011.5417C10.4138%2011.8472%2010.0777%2012%209.66661%2012H1.66661C1.2555%2012%200.919392%2011.8472%200.65828%2011.5417C0.397169%2011.2361%200.294391%2010.8833%200.349947%2010.4833L1.29995%203.81667C1.3555%203.49444%201.5055%203.22222%201.74995%203C1.99439%202.77778%202.28328%202.66667%202.61661%202.66667H3.78328C3.74995%202.55556%203.72217%202.44722%203.69995%202.34167C3.67772%202.23611%203.66661%202.12222%203.66661%202C3.66661%201.44444%203.86106%200.972222%204.24995%200.583333C4.63884%200.194444%205.11106%200%205.66661%200C6.22217%200%206.69439%200.194444%207.08328%200.583333C7.47217%200.972222%207.66661%201.44444%207.66661%202C7.66661%202.12222%207.6555%202.23611%207.63328%202.34167C7.61106%202.44722%207.58328%202.55556%207.54995%202.66667Z'%20fill='%236c7489'/%3e%3c/svg%3e",G={class:"target-ref"},j={key:3,class:"weight"},z=["width"],Z=T({__name:"TargetRef",props:{targetRef:{}},setup(k){const l=k,m=L(()=>{if(!l.targetRef.name)return null;switch(l.targetRef.kind){case"MeshService":case"MeshServiceSubset":return{name:"service-detail-view",params:{service:l.targetRef.name}};case"MeshGateway":return{name:"builtin-gateway-detail-view",params:{gateway:l.targetRef.name}};default:throw new Error(`Unsupported targetRef ${l.targetRef.kind}.`)}});return(r,g)=>{const o=w("XBadge"),t=w("XAction");return s(),a("span",G,[m.value!==null?(s(),C(t,{key:0,to:m.value},{default:c(()=>[_(o,null,{default:c(()=>[S(r.$slots,"default",{},void 0,!0)]),_:3})]),_:3},8,["to"])):(s(),C(o,{key:1},{default:c(()=>[S(r.$slots,"default",{},void 0,!0)]),_:3})),g[0]||(g[0]=e()),l.targetRef.kind==="MeshServiceSubset"&&l.targetRef.tags?(s(),C(X,{key:2,tags:l.targetRef.tags},null,8,["tags"])):p("",!0),g[1]||(g[1]=e()),l.targetRef.weight!==void 0&&l.targetRef.weight!==1?(s(),a("span",j,[n("img",{src:D,alt:"Weight",width:y(H)},null,8,z),e(" "+i(l.targetRef.weight),1)])):p("",!0)])}}}),q=B(Z,[["__scopeId","data-v-fb5c5475"]]),K={class:"filter"},O={key:0,class:"list"},U={class:"text-neutral"},J={key:1,class:"list"},Q={class:"text-neutral"},W={key:3,class:"list"},Y={class:"text-neutral"},tt={key:4,class:"list"},et={class:"text-neutral"},st=T({__name:"RuleFilter",props:{filter:{}},setup(k){const l=k;return(m,r)=>{const g=w("XBadge");return s(),a("div",K,[_(g,{appearance:"neutral"},{default:c(()=>[e(i(l.filter.type),1)]),_:1}),r[3]||(r[3]=e()),n("div",null,[l.filter.type==="RequestHeaderModifier"?(s(),a("div",O,[(s(!0),a(u,null,v(l.filter.requestHeaderModifier,(o,t)=>(s(),a(u,{key:t},[o?(s(!0),a(u,{key:0},v(o,(h,b)=>(s(),a("span",{key:b},[n("span",U,i(t)+`:
                `,1),e(" "+i(typeof h=="string"?h:`${h.name}:${h.value}`),1)]))),128)):p("",!0)],64))),128))])):l.filter.type==="ResponseHeaderModifier"?(s(),a("div",J,[(s(!0),a(u,null,v(l.filter.responseHeaderModifier,(o,t)=>(s(),a(u,{key:t},[o?(s(!0),a(u,{key:0},v(o,(h,b)=>(s(),a("span",{key:b},[n("span",Q,i(t)+`:
                `,1),e(" "+i(typeof h=="string"?h:`${h.name}:${h.value}`),1)]))),128)):p("",!0)],64))),128))])):l.filter.type==="RequestMirror"?(s(),a(u,{key:2},[_(q,{"target-ref":l.filter.requestMirror.backendRef},{default:c(()=>[e(i(l.filter.requestMirror.backendRef.name),1)]),_:1},8,["target-ref"]),r[0]||(r[0]=e()),l.filter.requestMirror.percentage?(s(),a(u,{key:0},[e(`
          (`+i(l.filter.requestMirror.percentage)+`%)
        `,1)],64)):p("",!0)],64)):l.filter.type==="RequestRedirect"?(s(),a("div",W,[(s(!0),a(u,null,v(l.filter.requestRedirect,(o,t)=>(s(),a("div",{key:t},[o?(s(),a(u,{key:0},[n("span",Y,i(t)+`:
              `,1),r[1]||(r[1]=e()),typeof o=="object"?(s(),a(u,{key:0},[e(i(o.type==="ReplaceFullPath"?o.replaceFullPath:o.replacePrefixMatch),1)],64)):(s(),a(u,{key:1},[e(i(o),1)],64))],64)):p("",!0)]))),128))])):(s(),a("div",tt,[(s(!0),a(u,null,v(l.filter.urlRewrite,(o,t)=>(s(),a("div",{key:t},[o?(s(),a(u,{key:0},[n("span",et,i(t)+`:
              `,1),r[2]||(r[2]=e()),typeof o=="object"?(s(),a(u,{key:0},[e(i(o.type==="ReplaceFullPath"?o.replaceFullPath:o.replacePrefixMatch),1)],64)):(s(),a(u,{key:1},[e(i(o),1)],64))],64)):p("",!0)]))),128))]))])])}}}),nt=B(st,[["__scopeId","data-v-9f3b1269"]]),at={class:"stack-small"},lt={key:0,class:"match"},rt={key:1,class:"match"},it={class:"match"},ot={class:"list"},dt={class:"list"},ut={key:0},ct=T({__name:"RuleMatch",props:{match:{}},setup(k){const l=k;return(m,r)=>{const g=w("XBadge");return s(),a("dl",at,[l.match.method?(s(),a("div",lt,[n("dt",null,[_(g,{appearance:"neutral"},{default:c(()=>r[0]||(r[0]=[e(`
            Method
          `)])),_:1})]),r[1]||(r[1]=e()),n("dd",null,[_(g,null,{default:c(()=>[e(i(l.match.method),1)]),_:1})])])):p("",!0),r[14]||(r[14]=e()),l.match.path?(s(),a("div",rt,[n("dt",null,[_(g,{appearance:"neutral"},{default:c(()=>r[2]||(r[2]=[e(`
            Path
          `)])),_:1})]),r[5]||(r[5]=e()),n("dd",null,[n("span",null,[r[3]||(r[3]=n("span",{class:"text-neutral"},"type:",-1)),e(" "+i(l.match.path.type),1)]),r[4]||(r[4]=e()),n("code",null,i(l.match.path.value),1)])])):p("",!0),r[15]||(r[15]=e()),l.match.queryParams&&l.match.queryParams.length>0?(s(!0),a(u,{key:2},v(l.match.queryParams,(o,t)=>(s(),a("div",{key:t},[n("div",it,[n("dt",null,[_(g,{appearance:"neutral"},{default:c(()=>r[6]||(r[6]=[e(`
              QueryParameter
            `)])),_:1})]),r[9]||(r[9]=e()),n("dd",null,[n("div",ot,[n("span",null,[r[7]||(r[7]=n("span",{class:"text-neutral"},"type:",-1)),e(" "+i(o.type),1)]),r[8]||(r[8]=e()),n("span",null,i(o.name)+":"+i(o.value),1)])])])]))),128)):p("",!0),r[16]||(r[16]=e()),l.match.headers&&l.match.headers.length>0?(s(!0),a(u,{key:3},v(l.match.headers,(o,t)=>(s(),a("div",{key:t,class:"match"},[n("dt",null,[_(g,{appearance:"neutral"},{default:c(()=>r[10]||(r[10]=[e(`
            Header
          `)])),_:1})]),r[13]||(r[13]=e()),n("dd",null,[n("div",dt,[n("span",null,[r[11]||(r[11]=n("span",{class:"text-neutral"},"type:",-1)),e(" "+i(o.type??"Exact"),1)]),r[12]||(r[12]=e()),n("span",null,[e(i(o.name),1),o.value?(s(),a("span",ut,":"+i(o.value),1)):p("",!0)])])])]))),128)):p("",!0)])}}}),pt=B(ct,[["__scopeId","data-v-512100d9"]]),mt={class:"listener-routes-card"},ft={class:"listener-routes"},yt={class:"column"},_t={class:"header"},gt={class:"title"},vt={class:"count"},ht={class:"content"},kt={class:"listener-list"},wt={class:"listener-card-header"},bt={key:0,class:"definition-list definition-list--horizontal mt-2"},Ct={key:0},$t={class:"text-neutral"},Rt={key:1},xt={class:"text-neutral"},Mt={class:"column"},Tt={class:"header"},Pt={class:"title"},Bt={class:"count"},qt={class:"content"},St={class:"to-rule-list"},Lt={class:"stack-with-borders"},Xt={class:"definition-list definition-list--horizontal mt-2"},Ht={class:"text-neutral visually-hidden"},It={key:0},Et={class:"text-neutral"},Nt={class:"text-neutral"},Vt={key:1},At={class:"text-neutral"},Ft={class:"list"},Dt={key:0},Gt={class:"stack-small mt-2"},jt={class:"stack-small mt-2"},zt={key:0},Zt={class:"stack-small mt-2"},Kt={key:1},Ot={class:"stack-small mt-2"},Ut={key:1},Jt={key:0,class:"stack-small mt-2"},Qt={class:"rule stack-small"},Wt={class:"stack-small mt-2"},Yt=T({__name:"ListenerRoutes",props:{meshGateway:{},selectedListenerIndex:{},policyTypesByName:{},inspectRules:{}},setup(k){const{t:l}=I(),m=k,r=L(()=>{const o=m.meshGateway.conf.listeners[m.selectedListenerIndex];return o?m.inspectRules.filter(t=>!!(t.ruleType==="to"&&F(o.tags??{},t.matchers)&&(t.type==="MeshHTTPRoute"&&["HTTP","HTTPS"].includes(o.protocol)||t.type==="MeshTCPRoute"&&["TCP","TLS"].includes(o.protocol)))):[]});function g(o){if(o.isTrusted){const t=o.currentTarget.querySelector("[data-action]");t instanceof HTMLElement&&typeof t.click=="function"&&t.click()}}return(o,t)=>{const h=w("XBadge"),b=w("XAction"),P=w("XEmptyState");return s(),a("div",mt,[n("div",ft,[n("div",yt,[n("div",_t,[n("h2",gt,i(y(l)("builtin-gateways.detail.listeners")),1),t[0]||(t[0]=e()),n("p",vt,i(o.meshGateway.conf.listeners.length),1)]),t[6]||(t[6]=e()),n("div",ht,[n("div",kt,[(s(!0),a(u,null,v(o.meshGateway.conf.listeners,(d,x)=>(s(),a("div",{key:x,class:E(["card listener-card",{active:x===m.selectedListenerIndex}]),"data-testid":"listener-card",onClick:g},[n("div",wt,[_(h,{appearance:"info"},{default:c(()=>[e(i(y(l)(`http.api.value.${d.protocol}`)),1)]),_:2},1024),t[1]||(t[1]=e()),_(b,{to:{name:"builtin-gateway-detail-view",query:{listener:x}},class:"listener-hostname","data-action":""},{default:c(()=>[e(i(d.hostname)+":"+i(d.port),1)]),_:2},1032,["to"])]),t[5]||(t[5]=e()),d.tags||d.tls?(s(),a("dl",bt,[d.tags?(s(),a("div",Ct,[n("dt",$t,i(y(l)("builtin-gateways.detail.tags"))+`:
                    `,1),t[2]||(t[2]=e()),n("dd",null,[_(X,{tags:d.tags},null,8,["tags"])])])):p("",!0),t[4]||(t[4]=e()),d.tls?(s(),a("div",Rt,[n("dt",xt,i(y(l)("http.api.property.tls"))+`:
                    `,1),t[3]||(t[3]=e()),n("dd",null,i(d.tls.mode??"TERMINATE"),1)])):p("",!0)])):p("",!0)],2))),128))])])]),t[22]||(t[22]=e()),n("div",Mt,[n("div",Tt,[n("h2",Pt,i(y(l)("builtin-gateways.detail.routes")),1),t[7]||(t[7]=e()),n("p",Bt,i(r.value.length),1)]),t[21]||(t[21]=e()),n("div",qt,[n("div",St,[r.value.length===0?(s(),C(P,{key:0},{default:c(()=>[e(i(y(l)("builtin-gateways.detail.no_rules",{listener:m.selectedListenerIndex+1})),1)]),_:1})):(s(!0),a(u,{key:1},v(r.value,(d,x)=>(s(),a("div",{key:x,class:"card route-card","data-testid":"route-card"},[n("div",Lt,[n("dl",Xt,[n("div",null,[n("dt",Ht,i(y(l)("builtin-gateways.detail.type"))+`:
                        `,1),t[8]||(t[8]=e()),n("dd",null,[_(h,null,{default:c(()=>[e(i(d.type),1)]),_:2},1024)])]),t[12]||(t[12]=e()),d.config.hostnames.includes("*")?p("",!0):(s(),a("div",It,[n("dt",Et,i(y(l)("builtin-gateways.detail.hostnames"))+`:
                          `,1),t[9]||(t[9]=e()),n("dd",null,i(d.config.hostnames.join(", ")),1)])),t[13]||(t[13]=e()),n("div",null,[n("dt",Nt,i(y(l)("builtin-gateways.detail.matchers"))+`:
                        `,1),t[10]||(t[10]=e()),n("dd",null,[d.matchers.length>0?(s(),C(A,{key:0,items:d.matchers},null,8,["items"])):(s(),a("code",Vt,"*"))])]),t[14]||(t[14]=e()),n("div",null,[n("dt",At,i(y(l)("builtin-gateways.detail.origins"))+`:
                        `,1),t[11]||(t[11]=e()),n("dd",null,[n("div",Ft,[(s(!0),a(u,null,v(d.origins,(f,M)=>(s(),C(h,{key:M},{default:c(()=>[_(b,{to:{name:"policy-detail-view",params:{mesh:f.mesh,policyPath:m.policyTypesByName[f.type].path,policy:f.name}}},{default:c(()=>[e(i(f.name),1)]),_:2},1032,["to"])]),_:2},1024))),128))])])])]),t[19]||(t[19]=e()),d.config.rules.length>0?(s(),a("div",Dt,[n("b",null,i(y(l)("builtin-gateways.detail.rules")),1),t[17]||(t[17]=e(`:

                      `)),n("div",Gt,[(s(!0),a(u,null,v(d.config.rules,(f,M)=>(s(),a("div",{key:M,class:"rule stack-small"},[n("div",null,[e(i(y(l)("builtin-gateways.detail.matches"))+`:

                            `,1),n("div",jt,[(s(!0),a(u,null,v(f.matches,($,R)=>(s(),C(pt,{key:R,match:$},null,8,["match"]))),128))])]),t[15]||(t[15]=e()),f.default.filters.length>0?(s(),a("div",zt,[e(i(y(l)("builtin-gateways.detail.filters"))+`:

                            `,1),n("div",Zt,[(s(!0),a(u,null,v(f.default.filters,($,R)=>(s(),C(nt,{key:R,filter:$},null,8,["filter"]))),128))])])):p("",!0),t[16]||(t[16]=e()),f.default.backendRefs.length>0?(s(),a("div",Kt,[e(i(y(l)("builtin-gateways.detail.services"))+`:

                            `,1),n("div",Ot,[(s(!0),a(u,null,v(f.default.backendRefs,($,R)=>(s(),a("div",{key:R},[_(q,{"target-ref":$},{default:c(()=>[e(i($.name),1)]),_:2},1032,["target-ref"])]))),128))])])):p("",!0)]))),128))])])):p("",!0),t[20]||(t[20]=e()),(f=>"default"in f.config)(d)?(s(),a("div",Ut,[n("b",null,i(y(l)("builtin-gateways.detail.default")),1),t[18]||(t[18]=e(`:

                      `)),d.config.default.backendRefs&&d.config.default.backendRefs.length>0?(s(),a("div",Jt,[n("div",Qt,[n("div",null,[e(i(y(l)("builtin-gateways.detail.services"))+`:

                            `,1),n("div",Wt,[(s(!0),a(u,null,v(d.config.default.backendRefs,(f,M)=>(s(),a("div",{key:M},[_(q,{"target-ref":f},{default:c(()=>[e(i(f.name),1)]),_:2},1032,["target-ref"])]))),128))])])])])):p("",!0)])):p("",!0)])]))),128))])])])])])}}}),te=B(Yt,[["__scopeId","data-v-092b3fa5"]]),ae=T({__name:"BuiltinGatewayDetailView",props:{gateway:{}},setup(k){const l=k;return(m,r)=>{const g=w("DataLoader"),o=w("DataSource"),t=w("AppView"),h=w("RouteView");return s(),C(h,{name:"builtin-gateway-detail-view",params:{mesh:"",gateway:"",listener:"0"}},{default:c(({route:b,uri:P})=>[_(t,null,{default:c(()=>[_(o,{src:P(y(N),"/policy-types",{})},{default:c(({data:d,error:x})=>[_(o,{src:P(y(V),"/meshes/:mesh/mesh-gateways/:name/rules",{mesh:b.params.mesh,name:b.params.gateway})},{default:c(({data:f,error:M})=>[_(g,{data:[d,f],errors:[x,M]},{default:c(()=>[f&&d?(s(),C(te,{key:0,"mesh-gateway":l.gateway,"selected-listener-index":Number(b.params.listener),"policy-types-by-name":d.policies.reduce(($,R)=>Object.assign($,{[R.name]:R}),{}),"inspect-rules":f.rules},null,8,["mesh-gateway","selected-listener-index","policy-types-by-name","inspect-rules"])):p("",!0)]),_:2},1032,["data","errors"])]),_:2},1032,["src"])]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});export{ae as default};
